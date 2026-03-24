<template>
  <section class="scoring">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">Contact Scoring</h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New Rule
        </b-button>
      </div>
    </header>

    <div class="columns">
      <div class="column is-8">
        <div class="box">
          <h3 class="title is-5">Scoring Rules</h3>
          <b-table :data="rules" :loading="loading" hoverable>
            <b-table-column v-slot="props" field="name" label="Rule" width="25%" :td-attrs="$utils.tdID">
              <a href="#" @click.prevent="editRule(props.row)">
                <strong>{{ props.row.name }}</strong>
              </a>
            </b-table-column>

            <b-table-column v-slot="props" field="event_type" label="Event" width="20%">
              <b-tag>{{ props.row.eventType }}</b-tag>
            </b-table-column>

            <b-table-column v-slot="props" field="score_value" label="Score" width="10%">
              <span :class="{ 'has-text-success': props.row.scoreValue > 0, 'has-text-danger': props.row.scoreValue < 0 }">
                {{ props.row.scoreValue > 0 ? '+' : '' }}{{ props.row.scoreValue }}
              </span>
            </b-table-column>

            <b-table-column v-slot="props" field="enabled" label="Status" width="10%">
              <b-tag :type="props.row.enabled ? 'is-success' : 'is-light'">
                {{ props.row.enabled ? 'Active' : 'Off' }}
              </b-tag>
            </b-table-column>

            <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
              <div>
                <a href="#" @click.prevent="editRule(props.row)" aria-label="Edit rule">
                  <b-icon icon="pencil-outline" size="is-small" />
                </a>
                <a href="#" @click.prevent="$utils.confirm(null, () => deleteRule(props.row))" aria-label="Delete rule">
                  <b-icon icon="trash-can-outline" size="is-small" />
                </a>
              </div>
            </b-table-column>

            <template #empty v-if="!loading">
              <empty-placeholder />
            </template>
          </b-table>
        </div>
      </div>

      <div class="column is-4">
        <div class="box">
          <h3 class="title is-5">Score Distribution</h3>
          <div class="score-summary">
            <p class="is-size-7 has-text-grey mb-3">
              Contact scores are updated automatically when events occur (opens, clicks, bounces, subscriptions).
              Inactive contacts have their scores decayed periodically.
            </p>
            <div class="score-events">
              <h4 class="title is-6">Event Types</h4>
              <div v-for="ev in eventTypes" :key="ev.value" class="mb-1 is-size-7">
                <b-tag size="is-small">{{ ev.value }}</b-tag> {{ ev.label }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Rule editor modal -->
    <b-modal v-model="isEditorVisible" :width="550" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Scoring Rule</p>
          <button type="button" class="delete" @click="isEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Rule name">
            <b-input v-model="form.name" placeholder="e.g. Email opened" required />
          </b-field>
          <b-field label="Event type">
            <b-select v-model="form.eventType" expanded>
              <option v-for="ev in eventTypes" :key="ev.value" :value="ev.value">{{ ev.label }}</option>
            </b-select>
          </b-field>
          <b-field label="Score change">
            <b-numberinput v-model="form.scoreValue" :min="-100" :max="100" />
          </b-field>
          <b-field label="Enabled">
            <b-switch v-model="form.enabled">{{ form.enabled ? 'Active' : 'Disabled' }}</b-switch>
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveRule" :loading="saving">Save</b-button>
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
      rules: [],
      loading: false,
      saving: false,

      isEditorVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),

      eventTypes: [
        { value: 'email.opened', label: 'Email opened' },
        { value: 'email.clicked', label: 'Link clicked' },
        { value: 'email.bounced', label: 'Email bounced' },
        { value: 'list.subscribed', label: 'List subscribed' },
        { value: 'list.unsubscribed', label: 'List unsubscribed' },
        { value: 'inactivity.30days', label: '30 days inactive' },
        { value: 'inactivity.60days', label: '60 days inactive' },
        { value: 'inactivity.90days', label: '90 days inactive' },
      ],
    };
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        eventType: 'email.opened',
        scoreValue: 5,
        enabled: true,
      };
    },

    getRules() {
      this.loading = true;
      this.$api.getScoringRules().then((data) => {
        this.rules = data.results || data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isEditorVisible = true;
    },

    editRule(r) {
      this.form = {
        name: r.name,
        eventType: r.eventType,
        scoreValue: r.scoreValue,
        enabled: r.enabled,
      };
      this.isEditing = true;
      this.editingId = r.id;
      this.isEditorVisible = true;
    },

    saveRule() {
      this.saving = true;
      const data = {
        name: this.form.name,
        event_type: this.form.eventType,
        score_value: this.form.scoreValue,
        enabled: this.form.enabled,
      };

      const fn = this.isEditing
        ? this.$api.updateScoringRule(this.editingId, data)
        : this.$api.createScoringRule(data);

      fn.then(() => {
        this.isEditorVisible = false;
        this.getRules();
        this.$utils.toast(this.isEditing ? 'Rule updated' : 'Rule created');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    deleteRule(r) {
      this.$api.deleteScoringRule(r.id).then(() => {
        this.getRules();
        this.$utils.toast(`Deleted "${r.name}"`);
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getRules);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getRules);
  },

  mounted() {
    this.getRules();
  },
});
</script>
