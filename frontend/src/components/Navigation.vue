<template>
  <b-menu-list>
    <b-menu-item :to="{ name: 'dashboard' }" tag="router-link" :active="activeItem.dashboard"
      icon="view-dashboard-variant-outline" :label="$t('menu.dashboard')" /><!-- dashboard -->

    <b-menu-item :expanded="activeGroup.lists" :active="activeGroup.lists" data-cy="lists"
      @update:active="(state) => toggleGroup('lists', state)" icon="format-list-bulleted-square"
      :label="$t('globals.terms.lists')">
      <b-menu-item :to="{ name: 'lists' }" tag="router-link" :active="activeItem.lists" data-cy="all-lists"
        icon="format-list-bulleted-square" :label="$t('menu.allLists')" />
      <b-menu-item :to="{ name: 'forms' }" tag="router-link" :active="activeItem.forms" class="forms"
        icon="newspaper-variant-outline" :label="$t('menu.forms')" />
    </b-menu-item><!-- lists -->

    <b-menu-item v-if="$can('subscribers:*')" :expanded="activeGroup.subscribers" :active="activeGroup.subscribers"
      data-cy="subscribers" @update:active="(state) => toggleGroup('subscribers', state)" icon="account-multiple"
      :label="$t('globals.terms.subscribers')">
      <b-menu-item v-if="$can('subscribers:get_all', 'subscribers:get')" :to="{ name: 'subscribers' }" tag="router-link"
        :active="activeItem.subscribers" data-cy="all-subscribers" icon="account-multiple"
        :label="$t('menu.allSubscribers')" />
      <b-menu-item v-if="$can('subscribers:import')" :to="{ name: 'import' }" tag="router-link"
        :active="activeItem.import" data-cy="import" icon="file-upload-outline" :label="$t('menu.import')" />
      <b-menu-item v-if="$can('bounces:get')" :to="{ name: 'bounces' }" tag="router-link" :active="activeItem.bounces"
        data-cy="bounces" icon="email-bounce" :label="$t('globals.terms.bounces')" />
    </b-menu-item><!-- subscribers -->

    <b-menu-item v-if="$can('campaigns:*')" :expanded="activeGroup.campaigns" :active="activeGroup.campaigns"
      data-cy="campaigns" @update:active="(state) => toggleGroup('campaigns', state)" icon="rocket-launch-outline"
      :label="$t('globals.terms.campaigns')">
      <b-menu-item v-if="$can('campaigns:get')" :to="{ name: 'campaigns' }" tag="router-link"
        :active="activeItem.campaigns" data-cy="all-campaigns" icon="rocket-launch-outline"
        :label="$t('menu.allCampaigns')" />
      <b-menu-item v-if="$can('campaigns:manage')" :to="{ name: 'campaign', params: { id: 'new' } }" tag="router-link"
        :active="activeItem.campaign" data-cy="new-campaign" icon="plus" :label="$t('menu.newCampaign')" />
      <b-menu-item v-if="$can('media:*')" :to="{ name: 'media' }" tag="router-link" :active="activeItem.media"
        data-cy="media" icon="image-outline" :label="$t('menu.media')" />
      <b-menu-item v-if="$can('templates:get')" :to="{ name: 'templates' }" tag="router-link"
        :active="activeItem.templates" data-cy="templates" icon="file-image-outline"
        :label="$t('globals.terms.templates')" />
      <b-menu-item v-if="$can('campaigns:get_analytics')" :to="{ name: 'campaignAnalytics' }" tag="router-link"
        :active="activeItem.campaignAnalytics" data-cy="analytics" icon="chart-bar"
        :label="$t('globals.terms.analytics')" />
      <b-menu-item v-if="$can('ab_tests:get')" :to="{ name: 'abTests' }" tag="router-link"
        :active="activeItem.abTests" data-cy="ab-tests" icon="ab-testing"
        label="A/B Tests" />
    </b-menu-item><!-- campaigns -->

    <b-menu-item v-if="$can('segments:*')" :expanded="activeGroup.segments" :active="activeGroup.segments"
      data-cy="segments" @update:active="(state) => toggleGroup('segments', state)" icon="segment"
      label="Segments">
      <b-menu-item v-if="$can('segments:get')" :to="{ name: 'segments' }" tag="router-link"
        :active="activeItem.segments" data-cy="all-segments" icon="segment"
        label="All Segments" />
    </b-menu-item><!-- segments -->

    <b-menu-item v-if="$can('drips:*')" :expanded="activeGroup.drips" :active="activeGroup.drips"
      data-cy="drips" @update:active="(state) => toggleGroup('drips', state)" icon="water-outline"
      label="Drip Campaigns">
      <b-menu-item v-if="$can('drips:get')" :to="{ name: 'drips' }" tag="router-link"
        :active="activeItem.drips" data-cy="all-drips" icon="water-outline"
        label="All Drips" />
    </b-menu-item><!-- drips -->

    <b-menu-item v-if="$can('automations:*')" :expanded="activeGroup.automations" :active="activeGroup.automations"
      data-cy="automations" @update:active="(state) => toggleGroup('automations', state)" icon="robot-outline"
      label="Automations">
      <b-menu-item v-if="$can('automations:get')" :to="{ name: 'automations' }" tag="router-link"
        :active="activeItem.automations" data-cy="all-automations" icon="robot-outline"
        label="All Automations" />
    </b-menu-item><!-- automations -->

    <b-menu-item v-if="$can('scoring:*')" :expanded="activeGroup.scoring" :active="activeGroup.scoring"
      data-cy="scoring" @update:active="(state) => toggleGroup('scoring', state)" icon="star-outline"
      label="Scoring">
      <b-menu-item v-if="$can('scoring:get')" :to="{ name: 'scoring' }" tag="router-link"
        :active="activeItem.scoring" data-cy="scoring-rules" icon="star-outline"
        label="Scoring Rules" />
    </b-menu-item><!-- scoring -->

    <b-menu-item v-if="$can('deals:*', 'activities:*')" :expanded="activeGroup.crm" :active="activeGroup.crm"
      data-cy="crm" @update:active="(state) => toggleGroup('crm', state)" icon="briefcase-outline"
      label="CRM">
      <b-menu-item v-if="$can('deals:get')" :to="{ name: 'deals' }" tag="router-link"
        :active="activeItem.deals" data-cy="deals" icon="briefcase-outline"
        label="Deals" />
      <b-menu-item v-if="$can('activities:get')" :to="{ name: 'activities' }" tag="router-link"
        :active="activeItem.activities" data-cy="activities" icon="format-list-bulleted-square"
        label="Activities" />
    </b-menu-item><!-- crm -->

    <b-menu-item v-if="$can('warming:*')" :expanded="activeGroup.warming" :active="activeGroup.warming"
      data-cy="warming" @update:active="(state) => toggleGroup('warming', state)" icon="fire"
      label="Email Warming">
      <b-menu-item v-if="$can('warming:get')" :to="{ name: 'warmingCampaigns' }" tag="router-link"
        :active="activeItem.warmingCampaigns" data-cy="warming-campaigns" icon="fire"
        label="All Campaigns" />
      <b-menu-item v-if="$can('warming:get')" :to="{ name: 'warmingRecipients' }" tag="router-link"
        :active="activeItem.warmingRecipients" data-cy="warming-recipients" icon="account-multiple"
        label="Recipients" />
      <b-menu-item v-if="$can('warming:get')" :to="{ name: 'warmingSenders' }" tag="router-link"
        :active="activeItem.warmingSenders" data-cy="warming-senders" icon="email-outline"
        label="Senders" />
      <b-menu-item v-if="$can('warming:get')" :to="{ name: 'warmingTemplates' }" tag="router-link"
        :active="activeItem.warmingTemplates" data-cy="warming-templates" icon="file-document-outline"
        label="Templates" />
      <b-menu-item v-if="$can('warming:get')" :to="{ name: 'warmingSendLog' }" tag="router-link"
        :active="activeItem.warmingSendLog" data-cy="warming-log" icon="format-list-bulleted-square"
        label="Send Log" />
    </b-menu-item><!-- warming -->

    <b-menu-item v-if="$can('webhooks:*')" :expanded="activeGroup.webhooks" :active="activeGroup.webhooks"
      data-cy="webhooks" @update:active="(state) => toggleGroup('webhooks', state)" icon="webhook"
      label="Webhooks">
      <b-menu-item v-if="$can('webhooks:get')" :to="{ name: 'webhooks' }" tag="router-link"
        :active="activeItem.webhooks" data-cy="all-webhooks" icon="webhook"
        label="All Webhooks" />
    </b-menu-item><!-- webhooks -->

    <b-menu-item v-if="$can('users:*', 'roles:*')" :expanded="activeGroup.users" :active="activeGroup.users"
      data-cy="users" @update:active="(state) => toggleGroup('users', state)" icon="account-multiple"
      :label="$t('globals.terms.users')">
      <b-menu-item v-if="$can('users:get')" :to="{ name: 'users' }" tag="router-link" :active="activeItem.users"
        data-cy="users" icon="account-multiple" :label="$t('globals.terms.users')" />
      <b-menu-item v-if="$can('roles:get')" :to="{ name: 'userRoles' }" tag="router-link" :active="activeItem.userRoles"
        data-cy="userRoles" icon="newspaper-variant-outline" :label="$t('users.userRoles')" />
      <b-menu-item v-if="$can('roles:get')" :to="{ name: 'listRoles' }" tag="router-link" :active="activeItem.listRoles"
        data-cy="listRoles" icon="format-list-bulleted-square" :label="$t('users.listRoles')" />
    </b-menu-item><!-- users -->

    <b-menu-item v-if="$can('settings:*')" :expanded="activeGroup.settings" :active="activeGroup.settings"
      data-cy="settings" @update:active="(state) => toggleGroup('settings', state)" icon="cog-outline"
      :label="$t('menu.settings')">
      <b-menu-item v-if="$can('settings:get')" :to="{ name: 'settings' }" tag="router-link"
        :active="activeItem.settings" data-cy="all-settings" icon="cog-outline" :label="$t('menu.settings')" />
      <b-menu-item v-if="$can('settings:maintain')" :to="{ name: 'maintenance' }" tag="router-link"
        :active="activeItem.maintenance" data-cy="maintenance" icon="wrench-outline" :label="$t('menu.maintenance')" />
      <b-menu-item v-if="$can('settings:get')" :to="{ name: 'logs' }" tag="router-link" :active="activeItem.logs"
        data-cy="logs" icon="format-list-bulleted-square" :label="$t('menu.logs')" />
    </b-menu-item><!-- settings -->
  </b-menu-list>
</template>

<script>
import { mapState } from 'vuex';

export default {
  name: 'Navigation',

  props: {
    activeItem: { type: Object, default: () => { } },
    activeGroup: { type: Object, default: () => { } },
    isMobile: Boolean,
  },

  methods: {
    toggleGroup(group, state) {
      this.$emit('toggleGroup', group, state);
    },

    doLogout() {
      this.$emit('doLogout');
    },
  },

  computed: {
    ...mapState(['profile']),
  },

  mounted() {
    // A hack to close the open accordion burger menu items on click.
    // Buefy does not have a way to do this.
    if (this.isMobile) {
      document.querySelectorAll('.navbar li a[href]').forEach((e) => {
        e.onclick = () => {
          document.querySelector('.navbar-burger').click();
        };
      });
    }
  },
};

</script>
