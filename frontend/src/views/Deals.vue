<template>
  <section class="deals">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Deals
          <span v-if="!isNaN(deals.total)">({{ deals.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New Deal
        </b-button>
      </div>
    </header>

    <!-- Pipeline view -->
    <b-tabs v-model="viewMode" type="is-toggle" size="is-small">
      <b-tab-item label="Table" value="table" />
      <b-tab-item label="Pipeline" value="pipeline" />
    </b-tabs>

    <!-- Pipeline (Kanban) view -->
    <div v-if="viewMode === 'pipeline'" class="pipeline-view columns is-multiline">
      <div v-for="stage in stages" :key="stage" class="column is-2">
        <div class="pipeline-column">
          <h4 class="title is-6 has-text-centered">{{ stage }}</h4>
          <p class="is-size-7 has-text-grey has-text-centered mb-3">
            {{ pipelineData[stage] ? pipelineData[stage].count : 0 }} deals
            &bull; ${{ pipelineData[stage] ? pipelineData[stage].totalValue.toFixed(0) : '0' }}
          </p>
          <div v-for="d in getDealsByStage(stage)" :key="d.id" class="deal-card box p-3 mb-2">
            <p><strong class="is-size-7">{{ d.name }}</strong></p>
            <p class="is-size-7 has-text-grey">${{ d.value }}</p>
            <p class="is-size-7 has-text-grey" v-if="d.subscriberEmail">{{ d.subscriberEmail }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Table view -->
    <b-table v-if="viewMode === 'table'" :data="deals.results" :loading="loading" hoverable paginated
      backend-pagination pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="deals.perPage" :total="deals.total" backend-sorting @sort="onSort">
      <b-table-column v-slot="props" field="name" label="Deal" width="20%" sortable :td-attrs="$utils.tdID">
        <a href="#" @click.prevent="editDeal(props.row)">
          <strong>{{ props.row.name }}</strong>
        </a>
      </b-table-column>

      <b-table-column v-slot="props" field="subscriber_id" label="Contact" width="18%">
        <router-link v-if="props.row.subscriberId" :to="{ name: 'subscriber', params: { id: props.row.subscriberId } }">
          {{ props.row.subscriberEmail || `#${props.row.subscriberId}` }}
        </router-link>
      </b-table-column>

      <b-table-column v-slot="props" field="value" label="Value" width="10%" sortable>
        {{ props.row.currency }} {{ props.row.value }}
      </b-table-column>

      <b-table-column v-slot="props" field="stage" label="Stage" width="10%">
        <b-tag>{{ props.row.stage || 'N/A' }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="status" label="Status" width="8%" sortable>
        <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="expected_close" label="Expected Close" width="12%">
        {{ props.row.expectedClose ? $utils.niceDate(props.row.expectedClose) : '-' }}
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="12%" sortable>
        {{ $utils.niceDate(props.row.createdAt) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="8%" align="right">
        <div>
          <a href="#" @click.prevent="editDeal(props.row)" aria-label="Edit deal">
            <b-icon icon="pencil-outline" size="is-small" />
          </a>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteDeal(props.row))" aria-label="Delete deal">
            <b-icon icon="trash-can-outline" size="is-small" />
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>

    <!-- Deal editor modal -->
    <b-modal v-model="isEditorVisible" :width="650" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Deal</p>
          <button type="button" class="delete" @click="isEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Deal name">
            <b-input v-model="form.name" placeholder="Deal name" required />
          </b-field>
          <div class="columns">
            <div class="column">
              <b-field label="Value">
                <b-numberinput v-model="form.value" :min="0" :step="0.01" />
              </b-field>
            </div>
            <div class="column is-4">
              <b-field label="Currency">
                <b-select v-model="form.currency" expanded>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                </b-select>
              </b-field>
            </div>
          </div>
          <b-field label="Subscriber ID">
            <b-numberinput v-model="form.subscriberId" :min="0" />
          </b-field>
          <div class="columns">
            <div class="column">
              <b-field label="Stage">
                <b-select v-model="form.stage" expanded>
                  <option v-for="s in stages" :key="s" :value="s">{{ s }}</option>
                </b-select>
              </b-field>
            </div>
            <div class="column">
              <b-field label="Status">
                <b-select v-model="form.status" expanded>
                  <option value="open">Open</option>
                  <option value="won">Won</option>
                  <option value="lost">Lost</option>
                </b-select>
              </b-field>
            </div>
          </div>
          <b-field label="Expected close date">
            <b-datepicker v-model="form.expectedClose" placeholder="Select date" />
          </b-field>
          <b-field label="Notes">
            <b-input v-model="form.notes" type="textarea" placeholder="Optional notes" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveDeal" :loading="saving">Save</b-button>
        </footer>
      </div>
    </b-modal>
  </section>
</template>

<script>
import Vue from 'vue';
import { mapState } from 'vuex';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

export default Vue.extend({
  components: {
    EmptyPlaceholder,
  },

  data() {
    return {
      deals: { results: [], total: 0, perPage: 20 },
      pipelineData: {},
      loading: false,
      saving: false,
      viewMode: 'table',

      queryParams: {
        page: 1,
        orderBy: 'created_at',
        order: 'desc',
      },

      stages: [],

      isEditorVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),
    };
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        value: 0,
        currency: 'USD',
        subscriberId: 0,
        stage: 'Lead',
        status: 'open',
        expectedClose: null,
        notes: '',
      };
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getDeals();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getDeals();
    },

    getDeals() {
      this.loading = true;
      this.$api.getDeals({
        page: this.queryParams.page,
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      }).then((data) => {
        this.deals = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    getPipeline() {
      this.$api.getDealPipeline().then((data) => {
        const pipeline = {};
        (data || []).forEach((entry) => {
          pipeline[entry.stage] = entry;
        });
        this.pipelineData = pipeline;
      });
    },

    getDealsByStage(stage) {
      return (this.deals.results || []).filter((d) => d.stage === stage);
    },

    showNewForm() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isEditorVisible = true;
    },

    editDeal(d) {
      this.form = {
        name: d.name,
        value: d.value,
        currency: d.currency || 'USD',
        subscriberId: d.subscriberId,
        stage: d.stage || 'Lead',
        status: d.status,
        expectedClose: d.expectedClose ? new Date(d.expectedClose) : null,
        notes: d.notes || '',
      };
      this.isEditing = true;
      this.editingId = d.id;
      this.isEditorVisible = true;
    },

    saveDeal() {
      this.saving = true;
      const data = {
        name: this.form.name,
        value: this.form.value,
        currency: this.form.currency,
        subscriber_id: this.form.subscriberId,
        stage: this.form.stage,
        status: this.form.status,
        expected_close: this.form.expectedClose,
        notes: this.form.notes,
      };

      const fn = this.isEditing
        ? this.$api.updateDeal(this.editingId, data)
        : this.$api.createDeal(data);

      fn.then(() => {
        this.isEditorVisible = false;
        this.getDeals();
        this.getPipeline();
        this.$utils.toast(this.isEditing ? 'Deal updated' : 'Deal created');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    deleteDeal(d) {
      this.$api.deleteDeal(d.id).then(() => {
        this.getDeals();
        this.getPipeline();
        this.$utils.toast(`Deleted "${d.name}"`);
      });
    },
  },

  computed: {
    ...mapState(['settings']),
  },

  created() {
    this.$root.$on('page.refresh', () => {
      this.getDeals();
      this.getPipeline();
    });
  },

  destroyed() {
    this.$root.$off('page.refresh');
  },

  mounted() {
    // Load stages and defaults from settings if available.
    const s = this.$store.state.settings;
    if (s && s['crm.deal_stages'] && s['crm.deal_stages'].length > 0) {
      this.stages = s['crm.deal_stages'];
    } else {
      this.stages = ['Lead', 'Qualified', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost'];
    }
    if (s && s['crm.default_deal_stage']) {
      this.form.stage = s['crm.default_deal_stage'];
    }
    if (s && s['crm.default_currency']) {
      this.form.currency = s['crm.default_currency'];
    }

    this.getDeals();
    this.getPipeline();
  },
});
</script>

<style scoped>
.pipeline-column {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 0.75rem;
  min-height: 300px;
}
.deal-card {
  cursor: pointer;
}
.deal-card:hover {
  border-color: #3273dc;
}
</style>
