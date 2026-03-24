<template>
  <section class="webhook-log">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Webhook Log
          <span v-if="webhookName"> &mdash; {{ webhookName }}</span>
          <span v-if="!isNaN(logs.total)">({{ logs.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button icon-left="arrow-left" tag="router-link" :to="{ name: 'webhooks' }">
          Back to Webhooks
        </b-button>
      </div>
    </header>

    <b-table :data="logs.results" :loading="loading" hoverable detailed show-detail-icon
      paginated backend-pagination pagination-position="both" @page-change="onPageChange"
      :current-page="queryParams.page" :per-page="logs.perPage" :total="logs.total">
      <b-table-column v-slot="props" field="event" label="Event" width="20%">
        <b-tag>{{ props.row.event }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="response_code" label="Status" width="10%">
        <b-tag :type="props.row.responseCode >= 200 && props.row.responseCode < 300 ? 'is-success' : 'is-danger'">
          {{ props.row.responseCode || 'N/A' }}
        </b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="attempt" label="Attempt" width="8%">
        {{ props.row.attempt }}
      </b-table-column>

      <b-table-column v-slot="props" field="error" label="Error" width="30%">
        <span v-if="props.row.error" class="has-text-danger is-size-7">{{ props.row.error }}</span>
        <span v-else class="has-text-grey">-</span>
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Time" width="15%">
        {{ $utils.niceDate(props.row.createdAt, true) }}
      </b-table-column>

      <template #detail="props">
        <div class="columns">
          <div class="column">
            <h4 class="title is-6">Payload</h4>
            <pre class="is-size-7">{{ JSON.stringify(props.row.payload, null, 2) }}</pre>
          </div>
          <div class="column" v-if="props.row.responseBody">
            <h4 class="title is-6">Response</h4>
            <pre class="is-size-7">{{ props.row.responseBody }}</pre>
          </div>
        </div>
      </template>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>
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
      logs: { results: [], total: 0, perPage: 50 },
      loading: false,
      webhookName: '',
      queryParams: { page: 1 },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getLogs();
    },

    getLogs() {
      this.loading = true;
      const id = parseInt(this.$route.params.id, 10);
      this.$api.getWebhookLog(id, { page: this.queryParams.page }).then((data) => {
        this.logs = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    getWebhook() {
      const id = parseInt(this.$route.params.id, 10);
      this.$api.getWebhook(id).then((data) => {
        this.webhookName = data.name;
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getLogs);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getLogs);
  },

  mounted() {
    this.getWebhook();
    this.getLogs();
  },
});
</script>
