<template>
  <section class="dashboard content">
    <header class="columns is-vcentered mb-2">
      <div class="column is-narrow">
        <h1 class="title is-5 mb-0">
          {{ $utils.niceDate(new Date()) }}
        </h1>
      </div>
      <div class="column">
        <!-- Solomon fork: date-range filter. Drives the four window-aware
             metric tiles, the Campaign Health "Sent" column, the Views/Clicks
             charts, and is passed via URL query params on every click-through
             so the destination page (CampaignAnalytics, Send Log) opens with
             the same filter pre-applied. -->
        <div class="buttons has-addons mb-0" data-cy="dashboard-date-presets">
          <b-button size="is-small" :type="dateRange.preset === 'today' ? 'is-primary' : 'is-light'"
            @click="applyPreset('today')">Today</b-button>
          <b-button size="is-small" :type="dateRange.preset === '7d' ? 'is-primary' : 'is-light'"
            @click="applyPreset('7d')">7d</b-button>
          <b-button size="is-small" :type="dateRange.preset === '15d' ? 'is-primary' : 'is-light'"
            @click="applyPreset('15d')">15d</b-button>
          <b-button size="is-small" :type="dateRange.preset === '30d' ? 'is-primary' : 'is-light'"
            @click="applyPreset('30d')">30d</b-button>
          <b-button size="is-small" :type="dateRange.preset === 'custom' ? 'is-primary' : 'is-light'"
            @click="applyPreset('custom')">Custom</b-button>
        </div>
        <div v-if="dateRange.preset === 'custom'" class="columns is-mobile mt-1">
          <div class="column is-narrow">
            <b-datepicker v-model="dateRange.from" placeholder="From"
              icon="calendar-today" size="is-small" :max-date="dateRange.to || new Date()"
              @input="refreshWindow" />
          </div>
          <div class="column is-narrow">
            <b-datepicker v-model="dateRange.to" placeholder="To (optional)"
              icon="calendar-today" size="is-small" :min-date="dateRange.from"
              :max-date="new Date()" @input="refreshWindow" />
          </div>
        </div>
      </div>
    </header>

    <!-- Solomon fork: four window-bound metric tiles summed across all
         running campaigns for the chosen date filter. Click the arrow icon
         to drill into Campaign Analytics with the same range pre-applied. -->
    <section class="wrap">
      <div class="tile is-ancestor">
        <div class="tile is-vertical is-12">
          <div class="tile">
            <div class="tile is-parent relative" v-for="m in metricTiles" :key="m.key">
              <b-loading v-if="isMetricsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" :data-cy="`metric-${m.key}`">
                <div class="columns is-mobile is-vcentered">
                  <div class="column">
                    <p class="title is-3 mb-0">
                      <b-icon :icon="m.icon" />
                      {{ m.value.toLocaleString() }}
                    </p>
                    <p class="is-size-7 has-text-grey mt-1">
                      {{ m.label }} <span class="has-text-weight-semibold">({{ windowLabel }})</span>
                    </p>
                  </div>
                  <div class="column is-narrow">
                    <b-button size="is-small" type="is-light" icon-right="arrow-top-right"
                      @click="goToAnalytics" :title="`Open Campaign Analytics for ${m.label}`" />
                  </div>
                </div>
              </article>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section class="counts wrap">
      <div class="tile is-ancestor">
        <div class="tile is-vertical is-12">
          <div class="tile">
            <div class="tile is-parent is-vertical relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="lists">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="format-list-bulleted-square" />
                      {{ $utils.niceNumber(counts.lists.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.list', counts.lists.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.public) }}</label>
                        {{ $t('lists.types.public') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.private) }}</label>
                        {{ $t('lists.types.private') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinSingle) }}</label>
                        {{ $t('lists.optins.single') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinDouble) }}</label>
                        {{ $t('lists.optins.double') }}
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- lists -->

              <article class="tile is-child notification" data-cy="campaigns">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="rocket-launch-outline" />
                      {{ $utils.niceNumber(counts.campaigns.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.campaign', counts.campaigns.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li v-for="(num, status) in counts.campaigns.byStatus" :key="status">
                        <label for="#" :data-cy="`campaigns-${status}`">{{ num }}</label>
                        {{ $t(`campaigns.status.${status}`) }}
                        <span v-if="status === 'running'" class="spinner is-tiny">
                          <b-loading :is-full-page="false" active />
                        </span>
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- campaigns -->
            </div><!-- block -->

            <div class="tile is-parent relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="subscribers">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="account-multiple" />
                      {{ $utils.niceNumber(counts.subscribers.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.subscriber', counts.subscribers.total) }}
                    </p>
                  </div>

                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.blocklisted) }}</label>
                        {{ $t('subscribers.status.blocklisted') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.orphans) }}</label>
                        {{ $t('dashboard.orphanSubs') }}
                      </li>
                    </ul>
                  </div><!-- subscriber breakdown -->
                </div><!-- subscriber columns -->
                <hr />
                <div class="columns" data-cy="messages">
                  <div class="column is-12">
                    <p class="title">
                      <b-icon icon="email-outline" />
                      {{ $utils.niceNumber(totalMessagesSent) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $t('dashboard.messagesSent') }}
                    </p>
                  </div>
                </div>
              </article><!-- subscribers -->
            </div>
          </div>
          <div class="tile">
            <div class="tile is-parent relative">
              <b-loading v-if="isFeaturesLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="features-left">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="water-outline" />
                      {{ features.drips ? features.drips.total : 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Drip Campaigns</p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ features.drips ? features.drips.active : 0 }}</label>
                        Active
                      </li>
                    </ul>
                  </div>
                </div>
                <hr />
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="robot-outline" />
                      {{ features.automations ? features.automations.total : 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Automations</p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ features.automations ? features.automations.active : 0 }}</label>
                        Active
                      </li>
                    </ul>
                  </div>
                </div>
                <hr />
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="filter-variant" />
                      {{ features.segments || 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Segments</p>
                  </div>
                </div>
              </article>
            </div>
            <div class="tile is-parent relative">
              <b-loading v-if="isFeaturesLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="features-right">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="star-outline" />
                      {{ features.scoring_rules || 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Scoring Rules</p>
                  </div>
                </div>
                <hr />
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="handshake-outline" />
                      {{ features.deals ? features.deals.total : 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Deals</p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ features.deals ? features.deals.open : 0 }}</label>
                        Open
                      </li>
                    </ul>
                  </div>
                </div>
                <hr />
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="webhook" />
                      {{ features.webhooks ? features.webhooks.total : 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Webhooks</p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ features.webhooks ? features.webhooks.active : 0 }}</label>
                        Active
                      </li>
                    </ul>
                  </div>
                </div>
                <hr />
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="fire" />
                      {{ features.warming ? features.warming.total_sent : 0 }}
                    </p>
                    <p class="is-size-6 has-text-grey">Warming Emails Sent</p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ features.warming ? features.warming.campaigns : 0 }}</label>
                        Campaigns
                        ({{ features.warming ? features.warming.active : 0 }} active)
                      </li>
                      <li>
                        <label for="#">{{ features.warming ? features.warming.sent_today : 0 }}</label>
                        Sent today
                      </li>
                      <li>
                        <label for="#">{{ features.warming ? features.warming.total_errors : 0 }}</label>
                        Errors
                      </li>
                      <li>
                        <router-link :to="{ name: 'warmingSendLog' }"
                          class="is-size-7">View send log &rarr;</router-link>
                      </li>
                    </ul>
                  </div>
                </div>
              </article>
            </div>
          </div><!-- features row -->

          <!-- Solomon fork: Campaign Health — surfaces running campaigns and
               flags any whose last send was > 2hr ago as STALLED. The whole
               point is to make a stuck rate-limiter / silent worker stall
               obvious the moment you log in, instead of discovering it days
               later from someone asking "where are the conversions?". -->
          <div class="tile is-parent relative" v-if="health.length > 0">
            <article class="tile is-child notification">
              <h3 class="title is-size-6">
                <b-icon icon="heart-pulse" /> Campaign Health
                <span v-if="anyStalled" class="tag is-danger ml-2">{{ stalledCount }} STALLED</span>
              </h3>
              <b-table :data="health" striped hoverable>
                <b-table-column field="status" label="" v-slot="props" width="40">
                  <b-tag v-if="props.row.stalled" type="is-danger">STALLED</b-tag>
                  <b-tag v-else-if="props.row.idle" type="is-warning is-light">idle</b-tag>
                  <b-tag v-else type="is-success is-light">sending</b-tag>
                </b-table-column>
                <b-table-column field="name" label="Campaign" v-slot="props">
                  <router-link :to="{ name: 'campaign', params: { id: props.row.id } }">{{ props.row.name }}</router-link>
                </b-table-column>
                <b-table-column field="sent" label="Sent / Queued" v-slot="props">
                  {{ props.row.sent.toLocaleString() }} / {{ (props.row.toSend || 0).toLocaleString() }}
                </b-table-column>
                <b-table-column field="sentInWindow" :label="`Sent (${windowLabel})`" v-slot="props" width="120">
                  <strong>{{ (props.row.sentInWindow || 0).toLocaleString() }}</strong>
                </b-table-column>
                <b-table-column field="lastSentAt" label="Last send" v-slot="props">
                  <span v-if="props.row.lastSentAt" :class="{ 'has-text-danger': props.row.stalled }">
                    {{ $utils.niceDate(props.row.lastSentAt, true) }}
                  </span>
                  <span v-else class="has-text-grey">never</span>
                </b-table-column>
                <b-table-column field="rate" label="Send rate" v-slot="props">
                  <span class="has-text-grey">{{ props.row.sendRate || 0 }}/min</span>
                </b-table-column>
                <b-table-column label="" v-slot="props" width="60">
                  <b-button size="is-small" type="is-light" icon-right="arrow-top-right"
                    @click="goToCampaignSendLog(props.row.id)"
                    :title="`Open ${props.row.name} Send Log for ${windowLabel}`" />
                </b-table-column>
              </b-table>
              <p v-if="anyStalled" class="mt-3 is-size-7 has-text-grey">
                A campaign is flagged STALLED when status='running' but the last send is &gt; 2 hours old.
                Open the campaign and try the <strong>Reset window</strong> button. If that doesn't help,
                pause then resume the campaign to spawn a fresh worker pipe.
              </p>
            </article>
          </div>

          <div class="tile is-parent relative">
            <b-loading v-if="isChartsLoading" active :is-full-page="false" />
            <article class="tile is-child notification charts">
              <div class="columns">
                <div class="column is-6">
                  <div class="is-flex is-justify-content-space-between is-align-items-center">
                    <h3 class="title is-size-6 mb-0">
                      {{ $t('dashboard.campaignViews') }} <span class="has-text-grey is-size-7">({{ windowLabel }})</span>
                    </h3>
                    <b-button size="is-small" type="is-light" icon-right="arrow-top-right"
                      @click="goToAnalytics" title="Open in Campaign Analytics" />
                  </div>
                  <br />
                  <chart type="line" v-if="campaignViews" :data="campaignViews" />
                </div>
                <div class="column is-6">
                  <div class="is-flex is-justify-content-space-between is-align-items-center">
                    <h3 class="title is-size-6 mb-0">
                      {{ $t('dashboard.linkClicks') }} <span class="has-text-grey is-size-7">({{ windowLabel }})</span>
                    </h3>
                    <b-button size="is-small" type="is-light" icon-right="arrow-top-right"
                      @click="goToAnalytics" title="Open in Campaign Analytics" />
                  </div>
                  <br />
                  <chart type="line" v-if="campaignClicks" :data="campaignClicks" />
                </div>
              </div>
            </article>
          </div>
        </div>
      </div><!-- tile block -->
      <p v-if="settings['app.cache_slow_queries']" class="has-text-grey">
        *{{ $t('globals.messages.slowQueriesCached') }}
        <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferer"
          class="has-text-grey">
          <b-icon icon="link-variant" /> {{ $t('globals.buttons.learnMore') }}
        </a>
      </p>
    </section>
  </section>
</template>

<script>
import dayjs from 'dayjs';
import Vue from 'vue';
import { mapState } from 'vuex';
import { colors } from '../constants';
import Chart from '../components/Chart.vue';

export default Vue.extend({
  components: {
    Chart,
  },

  data() {
    return {
      isChartsLoading: true,
      isCountsLoading: true,
      isFeaturesLoading: true,
      isMetricsLoading: false,
      campaignViews: null,
      campaignClicks: null,
      counts: {
        lists: {},
        subscribers: {},
        campaigns: {},
        messages: 0,
      },
      features: {},
      // Solomon fork: per-campaign health rows for the dashboard widget.
      // Each: { id, name, sent, to_send, last_sent_at, send_rate, stalled, idle }
      health: [],

      // Solomon fork: dashboard-wide date filter. Drives the four
      // window-aware metric tiles, the Campaign Health "Sent" column, the
      // Views/Clicks charts, and is passed through as URL query params on
      // every click-through to a deeper analytics page.
      dateRange: {
        preset: 'today',
        from: null,
        to: null,
      },
      // The four window-aware metric tiles. Each is the sum across all
      // currently-running campaigns for the chosen window.
      metrics: {
        sent: 0,
        opened: 0,
        clicked: 0,
        bounced: 0,
      },
    };
  },

  methods: {
    // Solomon fork: pick a date-range preset and recompute from/to.
    // Re-fires every date-sensitive load (metrics, health, charts).
    applyPreset(preset) {
      const now = new Date();
      const startOfDay = (d) => {
        const x = new Date(d);
        x.setHours(0, 0, 0, 0);
        return x;
      };
      let from = null;
      let to = null;
      switch (preset) {
        case 'today':
          from = startOfDay(now);
          break;
        case '7d':
          from = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          break;
        case '15d':
          from = new Date(now.getTime() - 15 * 24 * 60 * 60 * 1000);
          break;
        case '30d':
          from = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
          break;
        case 'custom':
          // Keep existing from/to if already set, else default to last 7 days.
          if (!this.dateRange.from) from = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          else from = this.dateRange.from;
          if (this.dateRange.to) to = this.dateRange.to;
          break;
        default:
          from = startOfDay(now);
          break;
      }
      this.dateRange = { preset, from, to };
      this.refreshWindow();
    },

    // Refetch every dashboard surface that depends on the date range.
    refreshWindow() {
      this.loadMetrics();
      this.loadCharts();
      this.loadHealth();
    },

    // Pull window-bound counters across all running campaigns.
    loadMetrics() {
      this.isMetricsLoading = true;
      this.$api.getCampaigns({ per_page: 100 }).then((res) => {
        const all = (res && res.results) || [];
        const ids = all.filter((c) => c.status === 'running').map((c) => c.id);
        if (ids.length === 0) {
          this.metrics = {
            sent: 0,
            opened: 0,
            clicked: 0,
            bounced: 0,
          };
          this.isMetricsLoading = false;
          return;
        }
        const fromIso = this.rangeIso.from;
        const toIso = this.rangeIso.to;
        // Sent: sum totalSent from per-campaign send-log stats.
        const sentP = Promise.all(ids.map((id) => this.$api
          .getCampaignSendLogStats(id, this.statsParams)
          .catch(() => null))).then((rows) => rows.reduce((s, r) => s + ((r && r.totalSent) || 0), 0));
        // Opens / Clicks / Bounces: one call each, summed across campaigns.
        const params = {
          id: ids,
          from: fromIso || undefined,
          to: toIso || undefined,
        };
        const sumCount = (data) => (Array.isArray(data) ? data.reduce((s, d) => s + (d.count || 0), 0) : 0);
        const opensP = this.$api.getCampaignViewCounts(params).then(sumCount).catch(() => 0);
        const clicksP = this.$api.getCampaignClickCounts(params).then(sumCount).catch(() => 0);
        const bouncesP = this.$api.getCampaignBounceCounts(params).then(sumCount).catch(() => 0);
        Promise.all([sentP, opensP, clicksP, bouncesP]).then(([sent, opened, clicked, bounced]) => {
          this.metrics = {
            sent,
            opened,
            clicked,
            bounced,
          };
          this.isMetricsLoading = false;
        });
      }).catch(() => { this.isMetricsLoading = false; });
    },

    // Re-pull the Views / Clicks charts using the date-aware analytics
    // endpoints (the legacy /api/dashboard/charts is hardcoded to 30 days).
    loadCharts() {
      this.isChartsLoading = true;
      this.$api.getCampaigns({ per_page: 100 }).then((res) => {
        const all = (res && res.results) || [];
        const ids = all.filter((c) => c.status === 'running').map((c) => c.id);
        if (ids.length === 0) {
          this.campaignViews = {};
          this.campaignClicks = {};
          this.isChartsLoading = false;
          return;
        }
        const params = {
          id: ids,
          from: this.rangeIso.from || undefined,
          to: this.rangeIso.to || undefined,
        };
        Promise.all([
          this.$api.getCampaignViewCounts(params).catch(() => []),
          this.$api.getCampaignClickCounts(params).catch(() => []),
        ]).then(([views, clicks]) => {
          this.campaignViews = this.makeChart((views || []).map((d) => ({
            date: d.timestamp,
            count: d.count,
          })));
          this.campaignClicks = this.makeChart((clicks || []).map((d) => ({
            date: d.timestamp,
            count: d.count,
          })));
          this.isChartsLoading = false;
        });
      }).catch(() => { this.isChartsLoading = false; });
    },

    // Click-through helpers. Both encode the active date range so the
    // destination page applies the same filter without manual clicks.
    goToAnalytics() {
      const ids = this.health.map((c) => c.id);
      const query = {};
      if (ids.length) query.id = ids.join(',');
      if (this.rangeUnix.from) query.from = this.rangeUnix.from;
      if (this.rangeUnix.to) query.to = this.rangeUnix.to;
      this.$router.push({ name: 'campaignAnalytics', query });
    },
    goToCampaignSendLog(id) {
      const query = { preset: this.dateRange.preset };
      if (this.rangeIso.from) query.from = this.rangeIso.from;
      if (this.rangeIso.to) query.to = this.rangeIso.to;
      this.$router.push({
        name: 'campaign',
        params: { id },
        query,
        hash: '#sendlog',
      });
    },

    fetchData() {
      this.isCountsLoading = true;

      // Lifetime aggregates (entity counts, not date-bound).
      this.$api.getDashboardCounts().then((data) => {
        this.counts = data;
        this.isCountsLoading = false;
      });

      this.isFeaturesLoading = true;
      this.$api.getDashboardFeatureCounts().then((data) => {
        this.features = data;
        this.isFeaturesLoading = false;
      });

      // Date-aware tiles, charts, Campaign Health — all driven by the
      // current dateRange preset (defaults to 'today' on first load).
      this.applyPreset(this.dateRange.preset || 'today');
    },

    // Solomon fork: build the Campaign Health rows. For every running campaign,
    // pull the send-log stats (last_sent_at) and the live send-rate, compute
    // stalled/idle flags, sort STALLED first so the operator sees the problem.
    loadHealth() {
      // Use proper $api exports (which wrap the module-local axios client).
      // Earlier versions called `this.$api.http.get(...)` directly, but `http`
      // was never exposed under $api — every call threw TypeError, was
      // silently caught, and the Health tile stayed hidden.
      this.$api.getCampaigns({ per_page: 100 })
        .then((res) => {
          const all = (res && res.results) || [];
          const running = all.filter((c) => c.status === 'running');
          if (running.length === 0) {
            this.health = [];
            return;
          }
          // Build base rows; we'll fill in lastSentAt + sendRate + today/7d
          // counts per campaign in parallel below.
          const rows = running.map((c) => ({
            id: c.id,
            name: c.name,
            sent: c.sent || 0,
            toSend: c.toSend || 0,
            lastSentAt: null,
            sendRate: 0,
            sentToday: 0,
            sent7d: 0,
            stalled: false,
            idle: false,
          }));

          // Fetch send-rate map (one call covers all running campaigns).
          this.$api.getCampaignStats().then((statsRes) => {
            const list = (statsRes && Array.isArray(statsRes) ? statsRes : (statsRes.results || statsRes || [])) || [];
            const rateById = {};
            list.forEach((s) => { rateById[s.id] = s.sendRate || 0; });
            rows.forEach((_, i) => { rows[i] = { ...rows[i], sendRate: rateById[rows[i].id] || 0 }; });
          }).catch(() => { /* non-fatal */ });

          // Fetch lifetime stats (for last_sent_at + stalled flag) and
          // window-bound stats (for the new "Sent (window)" column) per
          // campaign in parallel. The window mirrors the dashboard's
          // dateRange filter — operator changes the preset, this column
          // updates automatically.
          const STALL_THRESHOLD_MS = 2 * 60 * 60 * 1000; // 2 hours
          const now = new Date();
          Promise.all(rows.map((row, i) => Promise.all([
            this.$api.getCampaignSendLogStats(row.id, {}).catch(() => null),
            this.$api.getCampaignSendLogStats(row.id, this.statsParams).catch(() => null),
          ]).then(([sr, win]) => {
            const stats = sr || {};
            const lastSentAt = stats.lastSentAt || null;
            const stalled = lastSentAt
              ? (now.getTime() - new Date(lastSentAt).getTime()) > STALL_THRESHOLD_MS
              : false;
            const idle = !lastSentAt;
            rows[i] = {
              ...rows[i],
              lastSentAt,
              stalled,
              idle,
              sentInWindow: (win && win.totalSent) || 0,
            };
          }))).then(() => {
            // Stalled rows first, then idle, then sending.
            rows.sort((a, b) => (b.stalled - a.stalled) || (b.idle - a.idle));
            this.health = rows;
          });
        })
        .catch(() => { /* non-fatal — health tile just won't render */ });
    },

    makeChart(data) {
      if (data.length === 0) {
        return {};
      }
      return {
        labels: data.map((d) => dayjs(d.date).format('DD MMM')),
        datasets: [
          {
            data: [...data.map((d) => d.count)],
            borderColor: colors.primary,
            borderWidth: 2,
            pointHoverBorderWidth: 5,
            pointBorderWidth: 0.5,
          },
        ],
      };
    },
  },

  computed: {
    ...mapState(['settings']),
    dayjs() {
      return dayjs;
    },
    totalMessagesSent() {
      const campaignSent = this.counts.messages || 0;
      const warmingSent = this.features.warming
        ? this.features.warming.total_sent || 0 : 0;
      return campaignSent + warmingSent;
    },
    // Solomon fork: how many running campaigns are flagged stalled.
    stalledCount() {
      return this.health.filter((c) => c.stalled).length;
    },
    anyStalled() {
      return this.stalledCount > 0;
    },
    // Solomon fork: filter range converted to the formats different
    // downstream surfaces want. Send Log + send-log/stats want ISO-8601;
    // CampaignAnalytics wants Unix seconds; both want raw Date for the
    // datepicker.
    rangeIso() {
      return {
        from: this.dateRange.from ? this.dateRange.from.toISOString() : '',
        to: this.dateRange.to ? this.dateRange.to.toISOString() : '',
      };
    },
    rangeUnix() {
      return {
        from: this.dateRange.from ? Math.floor(this.dateRange.from.getTime() / 1000) : null,
        to: this.dateRange.to ? Math.floor(this.dateRange.to.getTime() / 1000) : null,
      };
    },
    // Reusable params object for getCampaignSendLogStats / getCampaign*Counts.
    statsParams() {
      const p = {};
      if (this.rangeIso.from) p.from = this.rangeIso.from;
      if (this.rangeIso.to) p.to = this.rangeIso.to;
      return p;
    },
    // Friendly label for the "Sent (window)" column header in the Health tile.
    windowLabel() {
      switch (this.dateRange.preset) {
        case 'today': return 'Today';
        case '7d': return 'Last 7d';
        case '15d': return 'Last 15d';
        case '30d': return 'Last 30d';
        case 'custom': return 'Custom';
        default: return 'Window';
      }
    },
    // The four metric tiles rendered above the lifetime aggregates.
    metricTiles() {
      return [
        {
          key: 'sent',
          label: 'Sent',
          icon: 'email-fast-outline',
          value: this.metrics.sent || 0,
        },
        {
          key: 'opened',
          label: 'Opened',
          icon: 'email-open-outline',
          value: this.metrics.opened || 0,
        },
        {
          key: 'clicked',
          label: 'Clicked',
          icon: 'cursor-default-click-outline',
          value: this.metrics.clicked || 0,
        },
        {
          key: 'bounced',
          label: 'Bounced',
          icon: 'email-alert-outline',
          value: this.metrics.bounced || 0,
        },
      ];
    },
  },

  created() {
    this.$root.$on('page.refresh', this.fetchData);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.fetchData);
  },

  mounted() {
    this.fetchData();
  },
});
</script>
