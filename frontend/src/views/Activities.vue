<template>
  <section class="activities">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Contact Activities
          <span v-if="!isNaN(activities.total)">({{ activities.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          Log Activity
        </b-button>
      </div>
    </header>

    <!-- Filter -->
    <div class="columns mb-3">
      <div class="column is-4">
        <b-field>
          <b-input v-model="filterSubId" placeholder="Filter by subscriber ID" type="number"
            @keyup.native.enter="getActivities" />
          <p class="control">
            <b-button type="is-primary" icon-left="magnify" @click="getActivities" />
          </p>
        </b-field>
      </div>
    </div>

    <b-table :data="activities.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="activities.perPage" :total="activities.total">
      <b-table-column v-slot="props" field="subscriber_id" label="Contact" width="18%" :td-attrs="$utils.tdID">
        <router-link :to="{ name: 'subscriber', params: { id: props.row.subscriberId } }">
          Subscriber #{{ props.row.subscriberId }}
        </router-link>
      </b-table-column>

      <b-table-column v-slot="props" field="activity_type" label="Type" width="12%">
        <b-tag :class="'type-' + props.row.activityType">{{ props.row.activityType }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="description" label="Description" width="30%">
        {{ props.row.description || '-' }}
      </b-table-column>

      <b-table-column v-slot="props" field="created_by_name" label="By" width="12%">
        {{ props.row.createdByName || 'System' }}
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Time" width="15%">
        {{ $utils.niceDate(props.row.createdAt, true) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="8%" align="right">
        <div>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteActivity(props.row))" aria-label="Delete">
            <b-icon icon="trash-can-outline" size="is-small" />
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>

    <!-- New activity modal -->
    <b-modal v-model="isEditorVisible" :width="500" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">Log Activity</p>
          <button type="button" class="delete" @click="isEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Subscriber ID">
            <b-numberinput v-model="form.subscriberId" :min="1" required />
          </b-field>
          <b-field label="Activity type">
            <b-select v-model="form.activityType" expanded>
              <option value="note">Note</option>
              <option value="call">Call</option>
              <option value="meeting">Meeting</option>
              <option value="email">Email</option>
              <option value="task">Task</option>
            </b-select>
          </b-field>
          <b-field label="Description">
            <b-input v-model="form.description" type="textarea" placeholder="Activity details" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveActivity" :loading="saving">Save</b-button>
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
      activities: { results: [], total: 0, perPage: 50 },
      loading: false,
      saving: false,
      filterSubId: '',

      queryParams: { page: 1 },

      isEditorVisible: false,
      form: {
        subscriberId: 1,
        activityType: 'note',
        description: '',
      },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getActivities();
    },

    getActivities() {
      this.loading = true;
      const params = { page: this.queryParams.page };
      if (this.filterSubId) {
        params.subscriber_id = parseInt(this.filterSubId, 10);
      }
      this.$api.getActivities(params).then((data) => {
        this.activities = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.form = {
        subscriberId: this.filterSubId ? parseInt(this.filterSubId, 10) : 1,
        activityType: 'note',
        description: '',
      };
      this.isEditorVisible = true;
    },

    saveActivity() {
      this.saving = true;
      this.$api.createActivity({
        subscriber_id: this.form.subscriberId,
        activity_type: this.form.activityType,
        description: this.form.description,
      }).then(() => {
        this.isEditorVisible = false;
        this.getActivities();
        this.$utils.toast('Activity logged');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    deleteActivity(a) {
      this.$api.deleteActivity(a.id).then(() => {
        this.getActivities();
        this.$utils.toast('Activity deleted');
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getActivities);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getActivities);
  },

  mounted() {
    if (this.$route.query.subscriber_id) {
      this.filterSubId = this.$route.query.subscriber_id;
    }
    this.getActivities();
  },
});
</script>
