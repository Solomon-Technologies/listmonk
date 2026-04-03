<template>
  <section class="warming-campaigns">
    <header class="columns page-header">
      <div class="column is-6">
        <h1 class="title is-4">
          Warming Campaigns
          <span v-if="campaigns.length > 0">({{ campaigns.length }})</span>
        </h1>
      </div>
      <div class="column has-text-right is-flex is-align-items-center
        is-justify-content-flex-end" style="gap: 1rem;">
        <div v-if="config" class="is-flex is-align-items-center" style="gap: 0.5rem;">
          <span class="is-size-7 has-text-grey">Warming</span>
          <b-switch v-model="config.is_active" size="is-small" type="is-success"
            @input="toggleGlobalWarming" />
        </div>
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="campaigns" :loading="loading" hoverable paginated
      :per-page="20" pagination-position="both">
      <b-table-column v-slot="props" field="status" label="Status" width="10%" sortable
        :td-attrs="$utils.tdID">
        <b-dropdown :triggers="['click']" role="list" position="is-bottom-left">
          <template #trigger>
            <b-tag :class="props.row.status" style="cursor: pointer;">
              {{ props.row.status }}
              <b-icon icon="menu-down" size="is-small" />
            </b-tag>
          </template>
          <b-dropdown-item v-if="props.row.status !== 'draft'" role="listitem"
            @click="changeStatus(props.row, 'draft')">
            Draft
          </b-dropdown-item>
          <b-dropdown-item v-if="props.row.status !== 'active'" role="listitem"
            @click="changeStatus(props.row, 'active')">
            Activate
          </b-dropdown-item>
          <b-dropdown-item v-if="props.row.status !== 'paused'" role="listitem"
            @click="changeStatus(props.row, 'paused')">
            Pause
          </b-dropdown-item>
        </b-dropdown>
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name" width="16%" sortable>
        <a href="#" @click.prevent="viewCampaign(props.row)">
          <strong>{{ props.row.name }}</strong>
        </a>
      </b-table-column>

      <b-table-column v-slot="props" field="brand" label="Brand" width="10%" sortable>
        {{ props.row.brand }}
      </b-table-column>

      <b-table-column v-slot="props" label="Sender" width="14%">
        <span v-if="getSenderEmail(props.row.sender_id)">
          {{ getSenderEmail(props.row.sender_id) }}
        </span>
        <span v-else class="has-text-grey">
          {{ (props.row.sender_domains || []).join(', ') }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" label="Day Cap" width="8%">
        <span v-if="getEffectiveCap(props.row) > 0">
          {{ getEffectiveCap(props.row) }}/day
        </span>
        <span v-else>
          {{ (props.row.sends_per_run || 0) * (props.row.runs_per_day || 0) }}/day
        </span>
      </b-table-column>

      <b-table-column v-slot="props" label="Warmup Day" width="7%">
        <span v-if="props.row.warmup_start_date">
          Day {{ getWarmupDay(props.row.warmup_start_date) }}
        </span>
        <span v-else class="has-text-grey">-</span>
      </b-table-column>

      <b-table-column v-slot="props" label="Schedule" width="12%">
        {{ (props.row.schedule_times || []).join(', ') }}
        <b-tag v-if="props.row.business_hours_only" size="is-small" type="is-info is-light"
          style="margin-left: 4px;">BH</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="10%" sortable>
        {{ $utils.niceDate(props.row.created_at) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
        <div>
          <a href="#" @click.prevent="editCampaign(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(
            null,
            () => deleteCampaign(props.row),
          )" aria-label="Delete">
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

    <!-- View / Stats modal -->
    <b-modal v-model="isViewModalVisible" :width="680" has-modal-card>
      <div class="modal-card" v-if="viewingCampaign">
        <header class="modal-card-head">
          <p class="modal-card-title">
            {{ viewingCampaign.name }}
            <b-tag :class="viewingCampaign.status" size="is-small"
              style="margin-left: 8px; vertical-align: middle;">
              {{ viewingCampaign.status }}
            </b-tag>
          </p>
          <button type="button" class="delete"
            @click="isViewModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <!-- Stats row -->
          <div class="columns is-mobile mb-4">
            <div class="column has-text-centered">
              <p class="heading">Sent Today</p>
              <p class="title is-4 has-text-success">
                {{ viewStats ? viewStats.sent_today : '-' }}
              </p>
            </div>
            <div class="column has-text-centered">
              <p class="heading">Errors Today</p>
              <p class="title is-4 has-text-danger">
                {{ viewStats ? viewStats.errors_today : '-' }}
              </p>
            </div>
            <div class="column has-text-centered">
              <p class="heading">Total Sent</p>
              <p class="title is-4">
                {{ viewStats ? viewStats.total_sent : '-' }}
              </p>
            </div>
            <div class="column has-text-centered">
              <p class="heading">Total Errors</p>
              <p class="title is-4 has-text-grey">
                {{ viewStats ? viewStats.total_errors : '-' }}
              </p>
            </div>
          </div>

          <hr />

          <!-- Campaign info -->
          <div class="columns is-mobile">
            <div class="column is-6">
              <p class="is-size-7 has-text-grey mb-1">Brand</p>
              <p>{{ viewingCampaign.brand || '-' }}</p>
            </div>
            <div class="column is-6">
              <p class="is-size-7 has-text-grey mb-1">Sender</p>
              <p v-if="getSenderEmail(viewingCampaign.sender_id)">
                {{ getSenderEmail(viewingCampaign.sender_id) }}
              </p>
              <p v-else>
                {{ (viewingCampaign.sender_domains || []).join(', ') || '-' }}
              </p>
            </div>
          </div>
          <div class="columns is-mobile">
            <div class="column is-4">
              <p class="is-size-7 has-text-grey mb-1">Warmup Day</p>
              <p v-if="viewingCampaign.warmup_start_date">
                Day {{ getWarmupDay(viewingCampaign.warmup_start_date) }}
              </p>
              <p v-else class="has-text-grey">Not started</p>
            </div>
            <div class="column is-4">
              <p class="is-size-7 has-text-grey mb-1">Day Cap</p>
              <p v-if="getEffectiveCap(viewingCampaign) > 0">
                {{ getEffectiveCap(viewingCampaign) }}/day
              </p>
              <p v-else>
                {{ (viewingCampaign.sends_per_run || 0)
                  * (viewingCampaign.runs_per_day || 0) }}/day
              </p>
            </div>
            <div class="column is-4">
              <p class="is-size-7 has-text-grey mb-1">Hourly Cap</p>
              <p>{{ viewingCampaign.hourly_cap || 'None' }}</p>
            </div>
          </div>
          <div class="columns is-mobile">
            <div class="column is-6">
              <p class="is-size-7 has-text-grey mb-1">Schedule (UTC)</p>
              <p>{{ (viewingCampaign.schedule_times || []).join(', ') }}</p>
            </div>
            <div class="column is-6">
              <p class="is-size-7 has-text-grey mb-1">Options</p>
              <p>
                {{ viewingCampaign.sends_per_run }} sends/run,
                {{ viewingCampaign.runs_per_day }} runs/day
                <span v-if="viewingCampaign.business_hours_only">, BH only</span>
              </p>
            </div>
          </div>

          <!-- Ramp schedule preview -->
          <div v-if="viewRampLimits.length > 0" class="mt-3">
            <p class="is-size-7 has-text-grey mb-2">Progressive Ramp</p>
            <div class="is-flex" style="gap: 0.5rem; flex-wrap: wrap;">
              <b-tag v-for="(dl, idx) in viewRampLimits" :key="idx"
                :type="isRampActive(dl) ? 'is-success is-light' : 'is-light'"
                size="is-small">
                Day {{ dl.day }}: {{ dl.max }}/day
              </b-tag>
            </div>
          </div>

          <hr />

          <!-- Recent sends -->
          <p class="is-size-7 has-text-grey mb-2">Recent Sends</p>
          <b-table :data="viewRecentSends" :loading="viewSendsLoading"
            hoverable narrowed size="is-small">
            <b-table-column v-slot="props" label="Sender" width="25%">
              <span class="is-size-7">{{ props.row.sender_email }}</span>
            </b-table-column>
            <b-table-column v-slot="props" label="Recipient" width="25%">
              <span class="is-size-7">{{ props.row.recipient_email }}</span>
            </b-table-column>
            <b-table-column v-slot="props" label="Subject" width="25%">
              <span class="is-size-7">{{ props.row.subject }}</span>
            </b-table-column>
            <b-table-column v-slot="props" label="Status" width="10%">
              <b-tag size="is-small"
                :type="props.row.status === 'sent' ? 'is-success' : 'is-danger'">
                {{ props.row.status }}
              </b-tag>
            </b-table-column>
            <b-table-column v-slot="props" label="Time" width="15%">
              <span class="is-size-7">
                {{ $utils.niceDate(props.row.sent_at, true) }}
              </span>
            </b-table-column>
            <template #empty v-if="!viewSendsLoading">
              <div class="has-text-centered has-text-grey is-size-7 py-3">
                No sends yet
              </div>
            </template>
          </b-table>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isViewModalVisible = false">Close</b-button>
          <b-button type="is-light" icon-left="pencil-outline"
            @click="isViewModalVisible = false; editCampaign(viewingCampaign)">
            Edit
          </b-button>
        </footer>
      </div>
    </b-modal>

    <!-- New / Edit modal -->
    <b-modal v-model="isModalVisible" :width="560" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">
            {{ isEditing ? 'Edit' : 'New' }} Warming Campaign
          </p>
          <button type="button" class="delete" @click="isModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Name" message="A descriptive name for this warming campaign">
            <b-input v-model="form.name" placeholder="Campaign name" required />
          </b-field>
          <b-field label="Brand" message="The brand this campaign warms for">
            <b-input v-model="form.brand" placeholder="Brand name" />
          </b-field>
          <b-field label="Status" message="Campaign status">
            <b-select v-model="form.status" expanded>
              <option value="draft">Draft</option>
              <option value="active">Active</option>
              <option value="paused">Paused</option>
            </b-select>
          </b-field>
          <b-field label="Sender"
            message="The specific sender email for this warming campaign">
            <b-select v-model="form.senderId" expanded
              placeholder="Select a sender...">
              <option :value="null">-- No specific sender (domain-based) --</option>
              <option v-for="s in warmingSenders" :key="s.id" :value="s.id">
                {{ s.email }} ({{ s.brand }})
              </option>
            </b-select>
          </b-field>
          <b-field label="SMTP Messenger"
            message="Which SMTP to use (e.g. email-resend, email-resend-rule27). Leave blank for default.">
            <b-input v-model="form.messenger" placeholder="email-resend" />
          </b-field>
          <b-field v-if="!form.senderId" label="Sender Domains"
            message="Fallback: match senders by domain (only if no sender selected)">
            <b-taginput
              v-model="form.senderDomains"
              :data="filteredDomains"
              autocomplete
              :allow-new="true"
              :open-on-focus="true"
              placeholder="Select or type a domain..."
              @typing="domainQuery = $event"
            />
          </b-field>
          <div class="columns">
            <div class="column">
              <b-field label="Sends per run" message="Emails per batch">
                <b-numberinput v-model="form.sendsPerRun" min="1" max="20"
                  controls-position="compact" />
              </b-field>
            </div>
            <div class="column">
              <b-field label="Runs per day" message="Batches fired per day">
                <b-numberinput v-model="form.runsPerDay" min="1" max="10"
                  controls-position="compact" />
              </b-field>
            </div>
          </div>
          <b-field label="Schedule times"
            message="HH:MM times when batches fire (comma-separated, UTC)">
            <b-input v-model="form.scheduleTimes"
              placeholder="10:00,14:00,18:00,21:00" />
          </b-field>
          <div class="columns">
            <div class="column">
              <b-field label="Min delay (seconds)"
                message="Minimum random pause between sends">
                <b-numberinput v-model="form.randomDelayMin" min="0" max="600"
                  controls-position="compact" />
              </b-field>
            </div>
            <div class="column">
              <b-field label="Max delay (seconds)"
                message="Maximum random pause between sends">
                <b-numberinput v-model="form.randomDelayMax" min="0" max="600"
                  controls-position="compact" />
              </b-field>
            </div>
          </div>
          <hr />
          <h3 class="is-size-6 has-text-weight-semibold mb-3">
            Progressive Ramp &amp; Limits
          </h3>
          <div class="columns">
            <div class="column">
              <b-field label="Hourly cap"
                message="Max sends per hour (0 = unlimited)">
                <b-numberinput v-model="form.hourlyCap" min="0" max="100"
                  controls-position="compact" />
              </b-field>
            </div>
            <div class="column">
              <b-field label="Business hours only"
                message="Only send Mon-Fri 9AM-6PM ET">
                <b-switch v-model="form.businessHoursOnly" type="is-success" />
              </b-field>
            </div>
          </div>
          <b-field label="Daily limits (progressive ramp)"
            message="JSON array of {day, max} objects. Day 1 = first active day.">
            <b-input v-model="form.dailyLimitsStr" type="textarea" rows="3"
              placeholder="[{day:1,max:5},{day:3,max:10},{day:7,max:50}]" />
          </b-field>
          <b-field v-if="isEditing && form.warmupStartDate" label="Warmup start date"
            message="Auto-set on first active run. Edit to reset.">
            <b-input v-model="form.warmupStartDate" type="date" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isModalVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveCampaign" :loading="saving">
            {{ isEditing ? 'Save' : 'Create' }}
          </b-button>
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
      campaigns: [],
      warmingSenders: [],
      loading: false,
      saving: false,

      // Edit modal.
      isModalVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),

      // View/stats modal.
      isViewModalVisible: false,
      viewingCampaign: null,
      viewStats: null,
      viewRecentSends: [],
      viewSendsLoading: false,
      viewRampLimits: [],

      // Global config.
      config: null,

      domainQuery: '',
    };
  },

  computed: {
    allDomains() {
      const domains = this.warmingSenders
        .map((s) => s.email.split('@')[1])
        .filter((d) => d);
      return [...new Set(domains)];
    },

    filteredDomains() {
      return this.allDomains.filter(
        (d) => !this.form.senderDomains.includes(d)
          && d.toLowerCase().includes((this.domainQuery || '').toLowerCase()),
      );
    },
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        brand: '',
        status: 'draft',
        senderId: null,
        senderDomains: [],
        sendsPerRun: 3,
        runsPerDay: 4,
        scheduleTimes: '10:00,14:00,18:00,21:00',
        randomDelayMin: 30,
        randomDelayMax: 120,
        hourlyCap: 0,
        businessHoursOnly: false,
        dailyLimitsStr: '[]',
        warmupStartDate: '',
        messenger: '',
      };
    },

    getSenderEmail(senderIdVal) {
      if (!senderIdVal) return null;
      const s = this.warmingSenders.find((x) => x.id === senderIdVal);
      return s ? s.email : null;
    },

    getCampaigns() {
      this.loading = true;
      this.$api.getWarmingCampaigns().then((data) => {
        this.campaigns = data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    getSenders() {
      this.$api.getWarmingSenders().then((data) => {
        this.warmingSenders = data || [];
      });
    },

    getConfig() {
      this.$api.getWarmingConfig().then((data) => {
        this.config = data;
      }).catch(() => {});
    },

    toggleGlobalWarming(val) {
      if (!this.config) return;
      this.$api.updateWarmingConfig({
        sends_per_run: this.config.sends_per_run,
        runs_per_day: this.config.runs_per_day,
        schedule_times: this.config.schedule_times,
        random_delay_min_s: this.config.random_delay_min_s,
        random_delay_max_s: this.config.random_delay_max_s,
        is_active: val,
      }).then(() => {
        this.$utils.toast(val ? 'Warming enabled' : 'Warming disabled');
      }).catch(() => {
        this.config.is_active = !val;
      });
    },

    // View modal — opens stats for a campaign.
    viewCampaign(row) {
      this.viewingCampaign = row;
      this.viewStats = null;
      this.viewRecentSends = [];
      this.viewSendsLoading = true;
      this.isViewModalVisible = true;

      // Parse ramp limits.
      try {
        const limits = row.daily_limits || [];
        this.viewRampLimits = Array.isArray(limits) ? limits : JSON.parse(limits);
      } catch (e) {
        this.viewRampLimits = [];
      }

      // Fetch stats and recent sends in parallel.
      this.$api.getWarmingCampaignStats(row.id).then((data) => {
        this.viewStats = data;
      }).catch(() => {});

      this.$api.getWarmingSendLog({
        campaign_id: row.id, limit: 10, offset: 0,
      }).then((data) => {
        this.viewRecentSends = data.results || [];
        this.viewSendsLoading = false;
      }).catch(() => { this.viewSendsLoading = false; });
    },

    isRampActive(dl) {
      if (!this.viewingCampaign || !this.viewingCampaign.warmup_start_date) {
        return false;
      }
      const dayNum = this.getWarmupDay(this.viewingCampaign.warmup_start_date);
      // Find the active bracket.
      let activeDl = null;
      for (let i = 0; i < this.viewRampLimits.length; i += 1) {
        if (dayNum <= this.viewRampLimits[i].day) {
          activeDl = this.viewRampLimits[i];
          break;
        }
        activeDl = this.viewRampLimits[i];
      }
      return activeDl && activeDl.day === dl.day && activeDl.max === dl.max;
    },

    showNewForm() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isModalVisible = true;
    },

    editCampaign(row) {
      this.form = {
        name: row.name,
        brand: row.brand,
        status: row.status || 'draft',
        senderId: row.sender_id || null,
        senderDomains: row.sender_domains || [],
        sendsPerRun: row.sends_per_run,
        runsPerDay: row.runs_per_day,
        scheduleTimes: (row.schedule_times || []).join(', '),
        randomDelayMin: row.random_delay_min_s || 30,
        randomDelayMax: row.random_delay_max_s || 120,
        hourlyCap: row.hourly_cap || 0,
        businessHoursOnly: row.business_hours_only || false,
        dailyLimitsStr: JSON.stringify(row.daily_limits || [], null, 2),
        warmupStartDate: row.warmup_start_date
          ? row.warmup_start_date.substring(0, 10) : '',
        messenger: row.messenger || '',
      };
      this.isEditing = true;
      this.editingId = row.id;
      this.isModalVisible = true;
    },

    saveCampaign() {
      this.saving = true;

      const scheduleTimes = typeof this.form.scheduleTimes === 'string'
        ? this.form.scheduleTimes.split(',').map((s) => s.trim()).filter((s) => s)
        : this.form.scheduleTimes;

      let dailyLimits = [];
      try {
        dailyLimits = JSON.parse(this.form.dailyLimitsStr || '[]');
      } catch (e) {
        this.$utils.toast('Invalid daily limits JSON', 'is-danger');
        this.saving = false;
        return;
      }

      const payload = {
        name: this.form.name,
        brand: this.form.brand,
        status: this.form.status || 'draft',
        sender_id: this.form.senderId || null,
        sender_domains: this.form.senderDomains,
        sends_per_run: this.form.sendsPerRun,
        runs_per_day: this.form.runsPerDay,
        schedule_times: scheduleTimes,
        random_delay_min_s: this.form.randomDelayMin,
        random_delay_max_s: this.form.randomDelayMax,
        hourly_cap: this.form.hourlyCap || 0,
        business_hours_only: this.form.businessHoursOnly || false,
        daily_limits: dailyLimits,
        warmup_start_date: this.form.warmupStartDate || null,
        messenger: this.form.messenger || '',
      };

      const fn = this.isEditing
        ? this.$api.updateWarmingCampaign(this.editingId, payload)
        : this.$api.createWarmingCampaign(payload);

      fn.then(() => {
        this.isModalVisible = false;
        this.saving = false;
        this.getCampaigns();
        this.$utils.toast(this.isEditing ? 'Campaign updated' : 'Campaign created');
      }).catch(() => { this.saving = false; });
    },

    changeStatus(row, status) {
      const payload = {
        name: row.name,
        brand: row.brand,
        sender_id: row.sender_id || null,
        sender_domains: row.sender_domains || [],
        status,
        sends_per_run: row.sends_per_run,
        runs_per_day: row.runs_per_day,
        schedule_times: row.schedule_times || [],
        random_delay_min_s: row.random_delay_min_s,
        random_delay_max_s: row.random_delay_max_s,
        hourly_cap: row.hourly_cap || 0,
        business_hours_only: row.business_hours_only || false,
        daily_limits: row.daily_limits || [],
        warmup_start_date: row.warmup_start_date
          ? row.warmup_start_date.substring(0, 10) : null,
        messenger: row.messenger || '',
      };
      this.$api.updateWarmingCampaign(row.id, payload).then(() => {
        this.getCampaigns();
        this.$utils.toast(`"${row.name}" status changed to ${status}`);
      });
    },

    getWarmupDay(startDate) {
      if (!startDate) return '-';
      const start = new Date(startDate);
      const now = new Date();
      const diffMs = now - start;
      return Math.floor(diffMs / (1000 * 60 * 60 * 24)) + 1;
    },

    getEffectiveCap(row) {
      const limits = row.daily_limits || [];
      if (!limits.length || !row.warmup_start_date) return 0;
      const dayNum = this.getWarmupDay(row.warmup_start_date);
      let cap = 0;
      for (let i = 0; i < limits.length; i += 1) {
        if (dayNum <= limits[i].day) {
          cap = limits[i].max;
          break;
        }
        cap = limits[i].max;
      }
      return cap;
    },

    deleteCampaign(row) {
      this.$api.deleteWarmingCampaign(row.id).then(() => {
        this.getCampaigns();
        this.$utils.toast(`Deleted "${row.name}"`);
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getCampaigns);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getCampaigns);
  },

  mounted() {
    this.getCampaigns();
    this.getSenders();
    this.getConfig();
  },
});
</script>
