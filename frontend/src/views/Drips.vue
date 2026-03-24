<template>
  <section class="drips">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Drip Campaigns
          <span v-if="!isNaN(drips.total)">({{ drips.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="drips.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="drips.perPage" :total="drips.total" backend-sorting @sort="onSort">
      <b-table-column v-slot="props" field="status" label="Status" width="10%" sortable :td-attrs="$utils.tdID">
        <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name" width="25%" sortable>
        <router-link :to="{ name: 'drip', params: { id: props.row.id } }">
          <strong>{{ props.row.name }}</strong>
        </router-link>
        <p class="is-size-7 has-text-grey" v-if="props.row.description">{{ props.row.description }}</p>
      </b-table-column>

      <b-table-column v-slot="props" field="trigger_type" label="Trigger" width="12%">
        <b-tag>{{ props.row.triggerType }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" label="Steps" width="8%">
        {{ props.row.stepsCount || 0 }}
      </b-table-column>

      <b-table-column v-slot="props" label="Entered" width="8%">
        {{ $utils.formatNumber(props.row.totalEntered || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" label="Completed" width="8%">
        {{ $utils.formatNumber(props.row.totalCompleted || 0) }}
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="14%" sortable>
        {{ $utils.niceDate(props.row.createdAt) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="12%" align="right">
        <div>
          <router-link :to="{ name: 'drip', params: { id: props.row.id } }">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </router-link>
          <a v-if="props.row.status === 'draft'" href="#" aria-label="Activate"
            @click.prevent="$utils.confirm('Activate this drip campaign?', () => changeStatus(props.row, 'active'))">
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
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteDrip(props.row))" aria-label="Delete">
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

    <!-- New drip modal -->
    <b-modal v-model="isNewVisible" :width="500" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">New Drip Campaign</p>
          <button type="button" class="delete" @click="isNewVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Name">
            <b-input v-model="newForm.name" placeholder="Campaign name" required />
          </b-field>
          <b-field label="Trigger type">
            <b-select v-model="newForm.triggerType" expanded>
              <option value="subscription">List subscription</option>
              <option value="segment_entry">Segment entry</option>
              <option value="tag_added">Tag added</option>
              <option value="date_field">Date field</option>
              <option value="manual">Manual enrollment</option>
            </b-select>
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isNewVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="createDrip" :loading="saving">Create</b-button>
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
      drips: { results: [], total: 0, perPage: 20 },
      loading: false,
      saving: false,

      queryParams: {
        page: 1,
        orderBy: 'created_at',
        order: 'desc',
      },

      isNewVisible: false,
      newForm: { name: '', triggerType: 'subscription' },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getDrips();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getDrips();
    },

    getDrips() {
      this.loading = true;
      this.$api.getDripCampaigns({
        page: this.queryParams.page,
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      }).then((data) => {
        this.drips = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.newForm = { name: '', triggerType: 'subscription' };
      this.isNewVisible = true;
    },

    createDrip() {
      this.saving = true;
      this.$api.createDripCampaign({
        name: this.newForm.name,
        trigger_type: this.newForm.triggerType,
      }).then((data) => {
        this.isNewVisible = false;
        this.saving = false;
        this.$router.push({ name: 'drip', params: { id: data.id } });
      }).catch(() => { this.saving = false; });
    },

    changeStatus(d, status) {
      this.$api.updateDripCampaign(d.id, { ...d, status }).then(() => {
        this.getDrips();
        this.$utils.toast(`Drip "${d.name}" status changed to ${status}`);
      });
    },

    deleteDrip(d) {
      this.$api.deleteDripCampaign(d.id).then(() => {
        this.getDrips();
        this.$utils.toast(`Deleted "${d.name}"`);
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getDrips);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getDrips);
  },

  mounted() {
    this.getDrips();
  },
});
</script>
