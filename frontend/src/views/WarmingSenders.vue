<template>
  <section class="warming-senders">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Warming Senders
          <span v-if="!isNaN(senders.length)">({{ senders.length }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewModal">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="senders" :loading="loading" hoverable>
      <b-table-column v-slot="props" field="email" label="Email">
        {{ props.row.email }}
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name">
        {{ props.row.name }}
      </b-table-column>

      <b-table-column v-slot="props" field="brand" label="Brand">
        <span :style="{ color: props.row.brand_color, fontWeight: 600 }">
          {{ props.row.brand }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" field="domain" label="Domain">
        {{ props.row.email.split('@')[1] }}
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
          <a href="#" @click.prevent="editSender(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(
            `Delete warming sender <b>${props.row.email}</b>?`,
            () => deleteSender(props.row),
          )" class="has-text-danger" aria-label="Delete">
            <b-icon icon="trash-can-outline" size="is-small" />
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>

    <!-- New / Edit modal -->
    <b-modal v-model="isModalVisible" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Warming Sender</p>
          <button type="button" class="delete" @click="isModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Email">
            <b-input v-model="form.email" type="email" required />
          </b-field>
          <b-field label="Display name">
            <b-input v-model="form.name" />
          </b-field>
          <b-field label="Brand name">
            <b-input v-model="form.brand" />
          </b-field>
          <b-field label="Brand URL">
            <b-input v-model="form.brand_url" placeholder="https://..." />
          </b-field>
          <b-field label="Brand color">
            <div class="is-flex is-align-items-center" style="gap: 0.75rem;">
              <input type="color" v-model="form.brand_color" aria-label="Brand color picker"
                style="width: 48px; height: 36px; padding: 0; border: 1px solid #dbdbdb; border-radius: 4px; cursor: pointer;" />
              <b-input v-model="form.brand_color" placeholder="#F2C94C"
                style="max-width: 120px;" />
            </div>
          </b-field>
          <b-field label="Active" v-if="isEditing">
            <b-switch v-model="form.is_active" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isModalVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveSender" :loading="saving">
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
      senders: [],
      loading: false,
      saving: false,
      isModalVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),
    };
  },

  methods: {
    getEmptyForm() {
      return {
        email: '',
        name: '',
        brand: '',
        brand_url: '',
        brand_color: '#F2C94C',
        is_active: true,
      };
    },

    getSenders() {
      this.loading = true;
      this.$api.getWarmingSenders().then((data) => {
        this.senders = data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewModal() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isModalVisible = true;
    },

    editSender(row) {
      this.form = {
        email: row.email,
        name: row.name,
        brand: row.brand,
        brand_url: row.brand_url,
        brand_color: row.brand_color || '#F2C94C',
        is_active: row.is_active,
      };
      this.isEditing = true;
      this.editingId = row.id;
      this.isModalVisible = true;
    },

    saveSender() {
      this.saving = true;

      const payload = {
        email: this.form.email,
        name: this.form.name,
        brand: this.form.brand,
        brand_url: this.form.brand_url,
        brand_color: this.form.brand_color,
      };

      if (this.isEditing) {
        payload.is_active = this.form.is_active;
      }

      const fn = this.isEditing
        ? this.$api.updateWarmingSender(this.editingId, payload)
        : this.$api.createWarmingSender(payload);

      fn.then(() => {
        this.$utils.toast(this.isEditing ? 'Sender updated' : 'Sender added');
        this.isModalVisible = false;
        this.saving = false;
        this.getSenders();
      }).catch(() => { this.saving = false; });
    },

    toggleActive(row, val) {
      this.$api.updateWarmingSender(row.id, {
        email: row.email,
        name: row.name,
        brand: row.brand,
        brand_url: row.brand_url,
        brand_color: row.brand_color,
        is_active: val,
      }).then(() => {
        this.getSenders();
      });
    },

    deleteSender(row) {
      this.$api.deleteWarmingSender(row.id).then(() => {
        this.$utils.toast('Sender deleted');
        this.getSenders();
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getSenders);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getSenders);
  },

  mounted() {
    this.getSenders();
  },
});
</script>
