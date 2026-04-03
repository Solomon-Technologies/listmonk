<template>
  <section class="warming-send-log">
    <header class="columns page-header">
      <div class="column is-6">
        <h1 class="title is-4">
          Warming Send Log
          <span v-if="!isNaN(logs.total)">({{ logs.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-field grouped position="is-right">
          <b-field label="Campaign" label-position="on-border">
            <b-select v-model="filterCampaignId" @input="onFilterChange" size="is-small">
              <option :value="0">All campaigns</option>
              <option v-for="c in campaigns" :key="c.id" :value="c.id">
                {{ c.name }}
              </option>
            </b-select>
          </b-field>
        </b-field>
      </div>
    </header>

    <b-table :data="logs.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="logs.perPage" :total="logs.total">
      <b-table-column v-slot="props" field="campaign_name" label="Campaign" width="14%">
        {{ props.row.campaign_name || '-' }}
      </b-table-column>

      <b-table-column v-slot="props" field="sender_email" label="Sender" width="18%">
        {{ props.row.sender_email }}
      </b-table-column>

      <b-table-column v-slot="props" field="recipient_email" label="Recipient" width="18%">
        {{ props.row.recipient_email }}
      </b-table-column>

      <b-table-column v-slot="props" field="subject" label="Subject" width="24%">
        {{ props.row.subject }}
      </b-table-column>

      <b-table-column v-slot="props" field="status" label="Status" width="10%">
        <b-tag :type="props.row.status === 'sent' ? 'is-success' : 'is-danger'" size="is-small">
          {{ props.row.status }}
        </b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="sent_at" label="Sent At" width="16%">
        {{ $utils.niceDate(props.row.sent_at, true) }}
      </b-table-column>

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
      campaigns: [],
      loading: false,
      queryParams: { page: 1 },
      filterCampaignId: 0,
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getLogs();
    },

    onFilterChange() {
      this.queryParams.page = 1;
      this.getLogs();
    },

    getCampaigns() {
      this.$api.getWarmingCampaigns().then((data) => {
        this.campaigns = data || [];
      });
    },

    getLogs() {
      this.loading = true;
      const offset = (this.queryParams.page - 1) * this.logs.perPage;
      const params = {
        offset,
        limit: this.logs.perPage,
      };
      if (this.filterCampaignId > 0) {
        params.campaign_id = this.filterCampaignId;
      }
      this.$api.getWarmingSendLog(params).then((data) => {
        this.logs.results = data.results || [];
        this.logs.total = data.total || 0;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getLogs);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getLogs);
  },

  mounted() {
    this.getCampaigns();
    this.getLogs();
  },
});
</script>
