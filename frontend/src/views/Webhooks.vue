<template>
  <section class="webhooks">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Webhooks
          <span v-if="!isNaN(webhooks.total)">({{ webhooks.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="webhooks.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="webhooks.perPage" :total="webhooks.total">
      <b-table-column v-slot="props" field="name" label="Name" :td-attrs="$utils.tdID">
        <a href="#" @click.prevent="editWebhook(props.row)">
          <strong>{{ props.row.name }}</strong>
        </a>
        <p class="is-size-7 has-text-grey">{{ props.row.url }}</p>
      </b-table-column>

      <b-table-column v-slot="props" field="enabled" label="Status" width="8%">
        <b-tag :type="props.row.enabled ? 'is-success' : 'is-light'">
          {{ props.row.enabled ? 'Active' : 'Disabled' }}
        </b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="events" label="Events" width="25%">
        <b-taglist>
          <b-tag class="is-small" v-for="e in props.row.events" :key="e">{{ e }}</b-tag>
        </b-taglist>
      </b-table-column>

      <b-table-column v-slot="props" field="total_sent" label="Sent" width="8%">
        {{ props.row.totalSent || 0 }}
      </b-table-column>

      <b-table-column v-slot="props" field="total_failed" label="Failed" width="8%">
        <span :class="{ 'has-text-danger': props.row.totalFailed > 0 }">
          {{ props.row.totalFailed || 0 }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="12%" align="right">
        <div>
          <a href="#" @click.prevent="editWebhook(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="testWebhook(props.row)" aria-label="Test">
            <b-tooltip label="Test" type="is-dark">
              <b-icon icon="play-outline" size="is-small" />
            </b-tooltip>
          </a>
          <router-link :to="{ name: 'webhookLog', params: { id: props.row.id } }" aria-label="View logs">
            <b-tooltip label="View logs" type="is-dark">
              <b-icon icon="format-list-bulleted-square" size="is-small" />
            </b-tooltip>
          </router-link>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteWebhook(props.row))" aria-label="Delete">
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

    <!-- Editor modal -->
    <b-modal v-model="isEditorVisible" :width="700" scroll="keep" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Webhook</p>
          <button type="button" class="delete" @click="isEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Name">
            <b-input v-model="form.name" placeholder="Webhook name" required />
          </b-field>
          <b-field label="URL">
            <b-input v-model="form.url" placeholder="https://..." type="url" required />
          </b-field>
          <b-field label="Secret (for HMAC signing)">
            <b-input v-model="form.secret" placeholder="Optional secret key" />
          </b-field>
          <b-field label="Enabled">
            <b-switch v-model="form.enabled">{{ form.enabled ? 'Active' : 'Disabled' }}</b-switch>
          </b-field>
          <b-field label="Max retries">
            <b-numberinput v-model="form.maxRetries" :min="0" :max="10" />
          </b-field>
          <b-field label="Timeout (seconds)">
            <b-numberinput v-model="form.timeoutSeconds" :min="1" :max="60" />
          </b-field>
          <hr />
          <b-field label="Events">
            <div class="columns is-multiline">
              <div class="column is-6" v-for="ev in availableEvents" :key="ev">
                <b-checkbox v-model="form.events" :native-value="ev">{{ ev }}</b-checkbox>
              </div>
            </div>
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveWebhook" :loading="saving">Save</b-button>
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
      webhooks: { results: [], total: 0, perPage: 20 },
      loading: false,
      saving: false,

      queryParams: { page: 1 },

      isEditorVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),

      availableEvents: [
        'subscriber.created', 'subscriber.updated', 'subscriber.optin',
        'subscriber.unsubscribed', 'campaign.started', 'campaign.finished',
        'bounce.received', 'drip.enrolled', 'drip.step_sent', 'drip.completed',
      ],
    };
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        url: '',
        secret: '',
        enabled: true,
        events: [],
        maxRetries: 3,
        timeoutSeconds: 10,
      };
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getWebhooks();
    },

    getWebhooks() {
      this.loading = true;
      this.$api.getWebhooks({ page: this.queryParams.page }).then((data) => {
        this.webhooks = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isEditorVisible = true;
    },

    editWebhook(w) {
      this.form = {
        name: w.name,
        url: w.url,
        secret: w.secret || '',
        enabled: w.enabled,
        events: [...(w.events || [])],
        maxRetries: w.maxRetries || 3,
        timeoutSeconds: w.timeoutSeconds || 10,
      };
      this.isEditing = true;
      this.editingId = w.id;
      this.isEditorVisible = true;
    },

    saveWebhook() {
      this.saving = true;
      const data = {
        name: this.form.name,
        url: this.form.url,
        secret: this.form.secret,
        enabled: this.form.enabled,
        events: this.form.events,
        max_retries: this.form.maxRetries,
        timeout_seconds: this.form.timeoutSeconds,
      };

      const fn = this.isEditing
        ? this.$api.updateWebhook(this.editingId, data)
        : this.$api.createWebhook(data);

      fn.then(() => {
        this.isEditorVisible = false;
        this.getWebhooks();
        this.$utils.toast(this.isEditing ? 'Webhook updated' : 'Webhook created');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    deleteWebhook(w) {
      this.$api.deleteWebhook(w.id).then(() => {
        this.getWebhooks();
        this.$utils.toast(`Deleted "${w.name}"`);
      });
    },

    testWebhook(w) {
      this.$api.testWebhook(w.id).then(() => {
        this.$utils.toast('Test webhook dispatched');
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getWebhooks);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getWebhooks);
  },

  mounted() {
    this.getWebhooks();
  },
});
</script>
