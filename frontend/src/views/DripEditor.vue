<template>
  <section class="drip-editor">
    <header class="columns page-header">
      <div class="column is-8">
        <h1 class="title is-4">
          <router-link :to="{ name: 'drips' }">Drip Campaigns</router-link>
          <span v-if="drip.name"> / {{ drip.name }}</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-tag :class="drip.status" size="is-medium">{{ drip.status }}</b-tag>
        <b-button v-if="drip.status === 'draft'" type="is-success" size="is-small" icon-left="play-outline"
          @click="$utils.confirm('Activate this drip campaign?', () => changeStatus('active'))">
          Activate
        </b-button>
        <b-button v-if="drip.status === 'active'" type="is-warning" size="is-small" icon-left="pause-circle-outline"
          @click="changeStatus('paused')">
          Pause
        </b-button>
        <b-button v-if="drip.status === 'paused'" type="is-success" size="is-small" icon-left="play-outline"
          @click="changeStatus('active')">
          Resume
        </b-button>
      </div>
    </header>

    <b-loading :is-full-page="false" v-model="loading" />

    <div class="columns" v-if="!loading">
      <!-- Settings panel -->
      <div class="column is-4">
        <div class="box">
          <h3 class="title is-5">Settings</h3>
          <b-field label="Name">
            <b-input v-model="drip.name" />
          </b-field>
          <b-field label="Description">
            <b-input v-model="drip.description" type="textarea" />
          </b-field>
          <b-field label="Trigger type">
            <b-select v-model="drip.triggerType" expanded>
              <option value="subscription">List subscription</option>
              <option value="segment_entry">Segment entry</option>
              <option value="tag_added">Tag added</option>
              <option value="date_field">Date field</option>
              <option value="manual">Manual enrollment</option>
            </b-select>
          </b-field>

          <!-- Trigger config based on type -->
          <div v-if="drip.triggerType === 'subscription'">
            <b-field label="Trigger lists">
              <b-taginput v-model="triggerListIds" placeholder="List ID" type="is-info" />
            </b-field>
          </div>
          <div v-if="drip.triggerType === 'tag_added'">
            <b-field label="Trigger tag">
              <b-input v-model="triggerTag" placeholder="Tag name" />
            </b-field>
          </div>

          <hr />
          <div class="fields stats">
            <p><span class="has-text-weight-semibold">Entered:</span> <span>{{ $utils.formatNumber(drip.totalEntered || 0) }}</span></p>
            <p><span class="has-text-weight-semibold">Completed:</span> <span>{{ $utils.formatNumber(drip.totalCompleted || 0) }}</span></p>
            <p><span class="has-text-weight-semibold">Exited:</span> <span>{{ $utils.formatNumber(drip.totalExited || 0) }}</span></p>
          </div>

          <hr />
          <b-button type="is-primary" expanded @click="saveDrip" :loading="saving">Save Settings</b-button>
        </div>
      </div>

      <!-- Steps timeline -->
      <div class="column is-8">
        <div class="box">
          <div class="columns">
            <div class="column">
              <h3 class="title is-5">Steps ({{ steps.length }})</h3>
            </div>
            <div class="column has-text-right">
              <b-button type="is-primary" size="is-small" icon-left="plus" @click="showNewStepForm">
                Add Step
              </b-button>
            </div>
          </div>

          <div v-if="steps.length === 0" class="has-text-centered has-text-grey py-5">
            No steps yet. Add the first step to your drip campaign.
          </div>

          <!-- Steps list -->
          <div v-for="(step, i) in steps" :key="step.id" class="drip-step">
            <div class="columns is-vcentered">
              <div class="column is-1 has-text-centered">
                <span class="step-number tag is-medium is-dark">{{ i + 1 }}</span>
              </div>
              <div class="column is-2">
                <span class="is-size-7 has-text-grey">Delay</span><br />
                <strong>{{ step.delayValue }} {{ step.delayUnit }}</strong>
              </div>
              <div class="column is-5">
                <strong>{{ step.name }}</strong>
                <p class="is-size-7 has-text-grey">{{ step.subject }}</p>
              </div>
              <div class="column is-2 has-text-centered">
                <div class="is-size-7">
                  <span>Sent: {{ step.sent || 0 }}</span><br />
                  <span>Opened: {{ step.opened || 0 }}</span><br />
                  <span>Clicked: {{ step.clicked || 0 }}</span>
                </div>
              </div>
              <div class="column is-2 has-text-right">
                <a href="#" @click.prevent="editStep(step)" aria-label="Edit step">
                  <b-icon icon="pencil-outline" size="is-small" />
                </a>
                <a href="#" @click.prevent="$utils.confirm(null, () => deleteStep(step))" aria-label="Delete step">
                  <b-icon icon="trash-can-outline" size="is-small" />
                </a>
              </div>
            </div>
            <div v-if="i < steps.length - 1" class="step-connector has-text-centered">
              <b-icon icon="arrow-down" size="is-small" class="has-text-grey-light" />
            </div>
          </div>
        </div>

        <!-- Enrollments -->
        <div class="box">
          <h3 class="title is-5">Enrollments</h3>
          <b-table :data="enrollments" :loading="enrollmentsLoading" hoverable>
            <b-table-column v-slot="props" field="email" label="Subscriber">
              <router-link :to="{ name: 'subscriber', params: { id: props.row.subscriberId } }">
                {{ props.row.email || `#${props.row.subscriberId}` }}
              </router-link>
            </b-table-column>
            <b-table-column v-slot="props" field="status" label="Status" width="12%">
              <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
            </b-table-column>
            <b-table-column v-slot="props" field="current_step" label="Current Step" width="15%">
              {{ props.row.currentStepId ? `Step #${getStepIndex(props.row.currentStepId)}` : '-' }}
            </b-table-column>
            <b-table-column v-slot="props" field="next_send_at" label="Next Send" width="18%">
              {{ props.row.nextSendAt ? $utils.niceDate(props.row.nextSendAt, true) : '-' }}
            </b-table-column>
            <b-table-column v-slot="props" field="entered_at" label="Entered" width="15%">
              {{ $utils.niceDate(props.row.enteredAt) }}
            </b-table-column>
            <template #empty v-if="!enrollmentsLoading">
              <div class="has-text-centered has-text-grey py-3">No enrollments yet</div>
            </template>
          </b-table>
        </div>
      </div>
    </div>

    <!-- Step editor modal -->
    <b-modal v-model="isStepEditorVisible" :width="800" scroll="keep" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditingStep ? 'Edit' : 'New' }} Step</p>
          <button type="button" class="delete" @click="isStepEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <div class="columns">
            <div class="column is-6">
              <b-field label="Step name">
                <b-input v-model="stepForm.name" placeholder="Welcome email" required />
              </b-field>
            </div>
            <div class="column is-3">
              <b-field label="Delay">
                <b-numberinput v-model="stepForm.delayValue" :min="0" />
              </b-field>
            </div>
            <div class="column is-3">
              <b-field label="Unit">
                <b-select v-model="stepForm.delayUnit" expanded>
                  <option value="minutes">Minutes</option>
                  <option value="hours">Hours</option>
                  <option value="days">Days</option>
                  <option value="weeks">Weeks</option>
                </b-select>
              </b-field>
            </div>
          </div>
          <b-field label="Subject line">
            <b-input v-model="stepForm.subject" placeholder="Email subject" required />
          </b-field>
          <b-field label="From email (optional)">
            <b-input v-model="stepForm.fromEmail" placeholder="Leave blank for default" />
          </b-field>
          <b-field label="Content type">
            <b-select v-model="stepForm.contentType" expanded>
              <option value="richtext">Rich text</option>
              <option value="html">Raw HTML</option>
              <option value="markdown">Markdown</option>
              <option value="plain">Plain text</option>
            </b-select>
          </b-field>
          <b-field label="Email body">
            <b-input v-model="stepForm.body" type="textarea" rows="12"
              placeholder="Email content (HTML, markdown, or plain text)" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isStepEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveStep" :loading="savingStep">Save</b-button>
        </footer>
      </div>
    </b-modal>
  </section>
</template>

<script>
import Vue from 'vue';

export default Vue.extend({
  data() {
    return {
      drip: {},
      steps: [],
      enrollments: [],
      loading: true,
      saving: false,
      savingStep: false,
      enrollmentsLoading: false,

      triggerListIds: [],
      triggerTag: '',

      isStepEditorVisible: false,
      isEditingStep: false,
      editingStepId: null,
      stepForm: this.getEmptyStepForm(),
    };
  },

  methods: {
    getEmptyStepForm() {
      return {
        name: '',
        subject: '',
        fromEmail: '',
        body: '',
        contentType: 'richtext',
        delayValue: 1,
        delayUnit: 'days',
        sequenceOrder: 0,
      };
    },

    getDrip() {
      const id = parseInt(this.$route.params.id, 10);
      this.loading = true;
      this.$api.getDripCampaign(id).then((data) => {
        this.drip = data;
        // Parse trigger config
        const cfg = data.triggerConfig || {};
        this.triggerListIds = (cfg.list_ids || []).map(String);
        this.triggerTag = cfg.tag || '';
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    getSteps() {
      const id = parseInt(this.$route.params.id, 10);
      this.$api.getDripSteps(id).then((data) => {
        this.steps = data || [];
      });
    },

    getEnrollments() {
      const id = parseInt(this.$route.params.id, 10);
      this.enrollmentsLoading = true;
      this.$api.getDripEnrollments(id, { per_page: 50 }).then((data) => {
        this.enrollments = data.results || data || [];
        this.enrollmentsLoading = false;
      }).catch(() => { this.enrollmentsLoading = false; });
    },

    saveDrip() {
      this.saving = true;
      const triggerConfig = {};
      if (this.drip.triggerType === 'subscription') {
        triggerConfig.list_ids = this.triggerListIds.map(Number);
      } else if (this.drip.triggerType === 'tag_added') {
        triggerConfig.tag = this.triggerTag;
      }

      this.$api.updateDripCampaign(this.drip.id, {
        name: this.drip.name,
        description: this.drip.description,
        trigger_type: this.drip.triggerType,
        trigger_config: triggerConfig,
        status: this.drip.status,
      }).then(() => {
        this.$utils.toast('Drip campaign saved');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    changeStatus(status) {
      this.$api.updateDripCampaign(this.drip.id, {
        ...this.drip,
        status,
      }).then(() => {
        this.drip.status = status;
        this.$utils.toast(`Status changed to ${status}`);
      });
    },

    showNewStepForm() {
      this.stepForm = this.getEmptyStepForm();
      this.stepForm.sequenceOrder = this.steps.length;
      this.isEditingStep = false;
      this.editingStepId = null;
      this.isStepEditorVisible = true;
    },

    editStep(step) {
      this.stepForm = {
        name: step.name,
        subject: step.subject,
        fromEmail: step.fromEmail || '',
        body: step.body || '',
        contentType: step.contentType || 'richtext',
        delayValue: step.delayValue,
        delayUnit: step.delayUnit,
        sequenceOrder: step.sequenceOrder,
      };
      this.isEditingStep = true;
      this.editingStepId = step.id;
      this.isStepEditorVisible = true;
    },

    saveStep() {
      this.savingStep = true;
      const dripId = parseInt(this.$route.params.id, 10);
      const data = {
        name: this.stepForm.name,
        subject: this.stepForm.subject,
        from_email: this.stepForm.fromEmail,
        body: this.stepForm.body,
        content_type: this.stepForm.contentType,
        delay_value: this.stepForm.delayValue,
        delay_unit: this.stepForm.delayUnit,
        sequence_order: this.stepForm.sequenceOrder,
      };

      const fn = this.isEditingStep
        ? this.$api.updateDripStep(dripId, this.editingStepId, data)
        : this.$api.createDripStep(dripId, data);

      fn.then(() => {
        this.isStepEditorVisible = false;
        this.getSteps();
        this.$utils.toast(this.isEditingStep ? 'Step updated' : 'Step created');
        this.savingStep = false;
      }).catch(() => { this.savingStep = false; });
    },

    deleteStep(step) {
      const dripId = parseInt(this.$route.params.id, 10);
      this.$api.deleteDripStep(dripId, step.id).then(() => {
        this.getSteps();
        this.$utils.toast('Step deleted');
      });
    },

    getStepIndex(stepId) {
      const idx = this.steps.findIndex((s) => s.id === stepId);
      return idx >= 0 ? idx + 1 : '?';
    },
  },

  created() {
    this.$root.$on('page.refresh', () => {
      this.getDrip();
      this.getSteps();
      this.getEnrollments();
    });
  },

  destroyed() {
    this.$root.$off('page.refresh');
  },

  mounted() {
    this.getDrip();
    this.getSteps();
    this.getEnrollments();
  },
});
</script>

<style scoped>
.drip-step {
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 0.75rem;
  margin-bottom: 0.5rem;
}
.step-connector {
  padding: 0.25rem 0;
}
.step-number {
  border-radius: 50%;
  width: 32px;
  height: 32px;
  line-height: 32px;
  text-align: center;
}
</style>
