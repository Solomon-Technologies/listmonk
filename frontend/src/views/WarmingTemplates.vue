<template>
  <section class="warming-templates">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Warming Templates
          <span v-if="!isNaN(templates.length)">({{ templates.length }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewModal">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="templates" :loading="loading" hoverable>
      <b-table-column v-slot="props" field="subject" label="Subject">
        <strong>{{ props.row.subject }}</strong>
      </b-table-column>

      <b-table-column v-slot="props" field="body" label="Body" width="35%">
        <span class="is-size-7 has-text-grey">
          {{ props.row.body.substring(0, 80) }}{{ props.row.body.length > 80 ? '...' : '' }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" field="is_active" label="Active" width="10%">
        <b-switch :value="props.row.is_active" size="is-small"
          @input="toggleActive(props.row, $event)" />
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="15%">
        {{ $utils.niceDate(props.row.created_at) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
        <div>
          <a href="#" @click.prevent="editTemplate(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(
            `Delete warming template <b>${props.row.subject}</b>?`,
            () => deleteTemplate(props.row),
          )" class="has-text-danger" aria-label="Delete">
            <b-tooltip label="Delete" type="is-dark">
              <b-icon icon="trash-can-outline" size="is-small" />
            </b-tooltip>
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>

    <!-- New / Edit modal -->
    <b-modal v-model="isModalVisible" has-modal-card>
      <div class="modal-card" style="width: 640px;">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Warming Template</p>
          <button type="button" class="delete" @click="isModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Subject">
            <b-input v-model="form.subject" placeholder="Quick sync" required />
          </b-field>
          <b-field label="Body (plain text)">
            <b-input v-model="form.body" type="textarea" rows="6"
              placeholder="Hey, just checking in..." />
          </b-field>
          <div class="notification is-info is-light is-size-7 mt-3">
            <strong>Available variables:</strong>
            <code v-pre>{{name}}</code> &mdash; recipient name,
            <code v-pre>{{date}}</code> &mdash; current date.
            Use in both subject and body.
          </div>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isModalVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveTemplate" :loading="saving">
            {{ isEditing ? 'Save' : 'Add' }}
          </b-button>
        </footer>
      </div>
    </b-modal>
  </section>
</template>

<script>
import Vue from 'vue';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

export default Vue.extend({
  components: {
    EmptyPlaceholder,
  },

  data() {
    return {
      templates: [],
      loading: false,
      saving: false,
      isModalVisible: false,
      isEditing: false,
      editingId: null,
      form: { subject: '', body: '' },
    };
  },

  methods: {
    getTemplates() {
      this.loading = true;
      this.$api.getWarmingTemplates().then((data) => {
        this.templates = data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewModal() {
      this.form = { subject: '', body: '' };
      this.isEditing = false;
      this.editingId = null;
      this.isModalVisible = true;
    },

    editTemplate(row) {
      this.form = {
        subject: row.subject,
        body: row.body,
      };
      this.isEditing = true;
      this.editingId = row.id;
      this.isModalVisible = true;
    },

    saveTemplate() {
      this.saving = true;

      const fn = this.isEditing
        ? this.$api.updateWarmingTemplate(this.editingId, {
          subject: this.form.subject,
          body: this.form.body,
          is_active: true,
        })
        : this.$api.createWarmingTemplate({
          subject: this.form.subject,
          body: this.form.body,
        });

      fn.then(() => {
        this.$utils.toast(this.isEditing ? 'Template updated' : 'Template added');
        this.isModalVisible = false;
        this.saving = false;
        this.getTemplates();
      }).catch(() => { this.saving = false; });
    },

    toggleActive(row, val) {
      this.$api.updateWarmingTemplate(row.id, {
        subject: row.subject,
        body: row.body,
        is_active: val,
      }).then(() => {
        this.getTemplates();
      });
    },

    deleteTemplate(row) {
      this.$api.deleteWarmingTemplate(row.id).then(() => {
        this.$utils.toast('Template deleted');
        this.getTemplates();
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getTemplates);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getTemplates);
  },

  mounted() {
    this.getTemplates();
  },
});
</script>
