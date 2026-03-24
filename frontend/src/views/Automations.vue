<template>
  <section class="automations">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Automations
          <span v-if="!isNaN(automations.total)">({{ automations.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="automations.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="automations.perPage" :total="automations.total">
      <b-table-column v-slot="props" field="status" label="Status" width="10%" :td-attrs="$utils.tdID">
        <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name" width="30%">
        <router-link :to="{ name: 'automation', params: { id: props.row.id } }">
          <strong>{{ props.row.name }}</strong>
        </router-link>
        <p class="is-size-7 has-text-grey" v-if="props.row.description">{{ props.row.description }}</p>
      </b-table-column>

      <b-table-column v-slot="props" label="Nodes" width="8%">
        {{ (props.row.canvas && props.row.canvas.nodes) ? props.row.canvas.nodes.length : 0 }}
      </b-table-column>

      <b-table-column v-slot="props" label="Entered" width="10%">
        {{ $utils.formatNumber(props.row.totalEntered || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" label="Completed" width="10%">
        {{ $utils.formatNumber(props.row.totalCompleted || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="14%">
        {{ $utils.niceDate(props.row.createdAt) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="12%" align="right">
        <div>
          <router-link :to="{ name: 'automation', params: { id: props.row.id } }">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </router-link>
          <a v-if="props.row.status === 'draft'" href="#" aria-label="Activate"
            @click.prevent="changeStatus(props.row, 'active')">
            <b-tooltip label="Activate" type="is-dark">
              <b-icon icon="play-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a v-if="props.row.status === 'active'" href="#" aria-label="Pause"
            @click.prevent="changeStatus(props.row, 'paused')">
            <b-tooltip label="Pause" type="is-dark">
              <b-icon icon="pause-circle-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a v-if="props.row.status === 'paused'" href="#" aria-label="Resume"
            @click.prevent="changeStatus(props.row, 'active')">
            <b-tooltip label="Resume" type="is-dark">
              <b-icon icon="play-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteAutomation(props.row))" aria-label="Delete">
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

    <!-- New automation modal -->
    <b-modal v-model="isNewVisible" :width="500" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">New Automation</p>
          <button type="button" class="delete" @click="isNewVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Name">
            <b-input v-model="newForm.name" placeholder="Automation name" required />
          </b-field>
          <b-field label="Description">
            <b-input v-model="newForm.description" type="textarea" placeholder="Optional description" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isNewVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="createAutomation" :loading="saving">Create</b-button>
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
      automations: { results: [], total: 0, perPage: 20 },
      loading: false,
      saving: false,
      queryParams: { page: 1 },
      isNewVisible: false,
      newForm: { name: '', description: '' },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getAutomations();
    },

    getAutomations() {
      this.loading = true;
      this.$api.getAutomations({ page: this.queryParams.page }).then((data) => {
        this.automations = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.newForm = { name: '', description: '' };
      this.isNewVisible = true;
    },

    createAutomation() {
      this.saving = true;
      this.$api.createAutomation({
        name: this.newForm.name,
        description: this.newForm.description,
      }).then((data) => {
        this.isNewVisible = false;
        this.saving = false;
        this.$router.push({ name: 'automation', params: { id: data.id } });
      }).catch(() => { this.saving = false; });
    },

    changeStatus(a, status) {
      this.$api.updateAutomation(a.id, { ...a, status }).then(() => {
        this.getAutomations();
        this.$utils.toast(`"${a.name}" status changed to ${status}`);
      });
    },

    deleteAutomation(a) {
      this.$api.deleteAutomation(a.id).then(() => {
        this.getAutomations();
        this.$utils.toast(`Deleted "${a.name}"`);
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getAutomations);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getAutomations);
  },

  mounted() {
    this.getAutomations();
  },
});
</script>
