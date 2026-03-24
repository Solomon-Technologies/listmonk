<template>
  <section class="ab-tests">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">A/B Tests</h1>
      </div>
    </header>

    <div class="notification is-info is-light">
      A/B tests are created from the Campaign editor. Go to a campaign and click the "A/B Test" tab to set up split testing.
    </div>

    <b-table :data="tests" :loading="loading" hoverable>
      <b-table-column v-slot="props" field="campaign" label="Campaign" width="25%">
        <router-link v-if="props.row.campaignId" :to="{ name: 'campaign', params: { id: props.row.campaignId } }">
          Campaign #{{ props.row.campaignId }}
        </router-link>
      </b-table-column>

      <b-table-column v-slot="props" field="test_type" label="Test Type" width="12%">
        <b-tag>{{ props.row.testType }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="status" label="Status" width="10%">
        <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="test_percentage" label="Test %" width="8%">
        {{ props.row.testPercentage }}%
      </b-table-column>

      <b-table-column v-slot="props" field="winner_metric" label="Winner By" width="10%">
        {{ props.row.winnerMetric }}
      </b-table-column>

      <b-table-column v-slot="props" label="Variants" width="20%">
        <div v-if="props.row.variants && props.row.variants.length" class="is-size-7">
          <div v-for="v in props.row.variants" :key="v.id" class="mb-1">
            <strong>{{ v.label }}:</strong>
            Open {{ (v.openRate || 0).toFixed(1) }}% / Click {{ (v.clickRate || 0).toFixed(1) }}%
            <b-tag v-if="props.row.winningVariantId === v.id" type="is-success" size="is-small">Winner</b-tag>
          </div>
        </div>
        <span v-else class="has-text-grey">No variants</span>
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="12%">
        {{ $utils.niceDate(props.row.createdAt) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="8%" align="right">
        <div>
          <a href="#" @click.prevent="viewTest(props.row)" aria-label="View details">
            <b-tooltip label="View details" type="is-dark">
              <b-icon icon="eye-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteTest(props.row))" aria-label="Delete">
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

    <!-- Detail modal -->
    <b-modal v-model="isDetailVisible" :width="700" has-modal-card>
      <div class="modal-card" v-if="selectedTest">
        <header class="modal-card-head">
          <p class="modal-card-title">A/B Test #{{ selectedTest.id }}</p>
          <button type="button" class="delete" @click="isDetailVisible = false" />
        </header>
        <section class="modal-card-body">
          <div class="columns">
            <div class="column">
              <p><strong>Type:</strong> {{ selectedTest.testType }}</p>
              <p><strong>Status:</strong> {{ selectedTest.status }}</p>
              <p><strong>Test size:</strong> {{ selectedTest.testPercentage }}%</p>
              <p><strong>Winner metric:</strong> {{ selectedTest.winnerMetric }}</p>
              <p><strong>Wait time:</strong> {{ selectedTest.winnerWaitHours }}h</p>
            </div>
          </div>
          <hr />
          <h4 class="title is-6">Variants</h4>
          <div v-for="v in (selectedTest.variants || [])" :key="v.id" class="box">
            <div class="columns">
              <div class="column">
                <strong>Variant {{ v.label }}</strong>
                <b-tag v-if="selectedTest.winningVariantId === v.id" type="is-success">Winner</b-tag>
              </div>
              <div class="column has-text-right is-size-7">
                Sent: {{ v.sent }} | Opened: {{ v.opened }} ({{ (v.openRate || 0).toFixed(1) }}%)
                | Clicked: {{ v.clicked }} ({{ (v.clickRate || 0).toFixed(1) }}%)
                | Bounced: {{ v.bounced }}
              </div>
            </div>
            <p v-if="v.subject" class="is-size-7"><strong>Subject:</strong> {{ v.subject }}</p>
            <p v-if="v.fromEmail" class="is-size-7"><strong>From:</strong> {{ v.fromEmail }}</p>
          </div>
        </section>
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
      tests: [],
      loading: false,
      isDetailVisible: false,
      selectedTest: null,
    };
  },

  methods: {
    getTests() {
      this.loading = true;
      this.$api.getABTests().then((data) => {
        this.tests = data.results || data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    viewTest(t) {
      this.$api.getABTest(t.id).then((data) => {
        this.selectedTest = data;
        this.isDetailVisible = true;
      });
    },

    deleteTest(t) {
      this.$api.deleteABTest(t.id).then(() => {
        this.getTests();
        this.$utils.toast('A/B test deleted');
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getTests);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getTests);
  },

  mounted() {
    this.getTests();
  },
});
</script>
