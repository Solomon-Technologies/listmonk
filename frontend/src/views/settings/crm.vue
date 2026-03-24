<template>
  <div class="items">
    <h2 class="is-size-4 mb-5">Deal Pipeline</h2>

    <b-field label="Pipeline Stages" label-position="on-border"
      message="Define the stages a deal moves through. Drag to reorder.">
      <b-taginput v-model="data['crm.deal_stages']"
        placeholder="Add a stage..." />
    </b-field>

    <div class="columns">
      <div class="column is-6">
        <b-field label="Default Stage for New Deals" label-position="on-border">
          <b-select v-model="data['crm.default_deal_stage']" expanded>
            <option v-for="stage in data['crm.deal_stages']" :key="stage" :value="stage">
              {{ stage }}
            </option>
          </b-select>
        </b-field>
      </div>
      <div class="column is-6">
        <b-field label="Default Currency" label-position="on-border">
          <b-select v-model="data['crm.default_currency']" expanded>
            <option v-for="cur in data['crm.currencies']" :key="cur" :value="cur">
              {{ cur }}
            </option>
          </b-select>
        </b-field>
      </div>
    </div>

    <hr />
    <h2 class="is-size-4 mb-5">Currencies</h2>

    <b-field label="Supported Currencies" label-position="on-border"
      message="Currency codes available when creating deals.">
      <b-taginput v-model="data['crm.currencies']"
        placeholder="Add a currency code (e.g. USD)..." />
    </b-field>

    <hr />
    <h2 class="is-size-4 mb-5">Activity Types</h2>

    <b-field label="Activity Types" label-position="on-border"
      message="Types of activities that can be logged on contacts.">
      <b-taginput v-model="data['crm.activity_types']"
        placeholder="Add an activity type..." />
    </b-field>
  </div>
</template>

<script>
import Vue from 'vue';

export default Vue.extend({
  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  data() {
    return {
      data: this.form,
    };
  },

  mounted() {
    // Ensure defaults exist if settings haven't been migrated yet.
    if (!this.data['crm.deal_stages']) {
      this.$set(this.data, 'crm.deal_stages', ['Lead', 'Qualified', 'Proposal', 'Negotiation', 'Closed Won', 'Closed Lost']);
    }
    if (!this.data['crm.currencies']) {
      this.$set(this.data, 'crm.currencies', ['USD', 'EUR', 'GBP']);
    }
    if (!this.data['crm.activity_types']) {
      this.$set(this.data, 'crm.activity_types', ['note', 'call', 'meeting', 'email', 'task']);
    }
    if (!this.data['crm.default_deal_stage']) {
      this.$set(this.data, 'crm.default_deal_stage', 'Lead');
    }
    if (!this.data['crm.default_currency']) {
      this.$set(this.data, 'crm.default_currency', 'USD');
    }
  },
});
</script>
