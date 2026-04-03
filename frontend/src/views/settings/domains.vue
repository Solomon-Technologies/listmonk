<template>
  <div class="items">
    <!-- Domains summary -->
    <div class="is-flex is-align-items-center is-justify-content-space-between mb-4">
      <h2 class="is-size-4">Domains</h2>
      <b-button type="is-primary" size="is-small" icon-left="plus" @click="showNewModal">
        Add Sender
      </b-button>
    </div>
    <p class="has-text-grey is-size-7 mb-4">
      Domains are extracted from sender emails. Add a sender to register a new domain.
    </p>
    <div class="columns is-multiline mb-5" v-if="domains.length > 0">
      <div class="column is-4" v-for="d in domains" :key="d.domain">
        <div class="box" style="padding: 1rem;">
          <div class="is-flex is-align-items-center is-justify-content-space-between">
            <div>
              <strong class="is-size-6">{{ d.domain }}</strong>
              <p class="is-size-7 has-text-grey">
                {{ d.activeCount }}/{{ d.totalCount }} sender{{ d.totalCount !== 1 ? 's' : '' }} active
              </p>
            </div>
            <b-tag :type="d.activeCount > 0 ? 'is-success' : 'is-light'" size="is-small">
              {{ d.activeCount > 0 ? 'Active' : 'Inactive' }}
            </b-tag>
          </div>
        </div>
      </div>
    </div>
    <div v-else-if="!loading" class="has-text-grey has-text-centered py-4 mb-5">
      No senders configured yet. Add a sender below to get started.
    </div>

    <hr />

    <!-- Senders table -->
    <div class="is-flex is-align-items-center is-justify-content-space-between mb-4">
      <h2 class="is-size-4">Senders</h2>
      <b-button type="is-primary" size="is-small" icon-left="plus" @click="showNewModal">
        Add Sender
      </b-button>
    </div>
    <p class="has-text-grey is-size-7 mb-4">
      Verified sender identities used for warming, drips, and automations.
    </p>

    <b-table :data="senders" :loading="loading" hoverable>
      <b-table-column v-slot="props" field="email" label="Email" width="22%">
        <strong>{{ props.row.email }}</strong>
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Display Name" width="15%">
        {{ props.row.name }}
      </b-table-column>

      <b-table-column v-slot="props" field="brand" label="Brand" width="15%">
        <span :style="{ color: props.row.brand_color, fontWeight: 600 }">
          {{ props.row.brand }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" label="Domain" width="15%">
        {{ props.row.email.split('@')[1] }}
      </b-table-column>

      <b-table-column v-slot="props" label="Color" width="8%">
        <span class="color-swatch"
          :style="{ backgroundColor: props.row.brand_color || '#ccc' }" />
      </b-table-column>

      <b-table-column v-slot="props" field="is_active" label="Active" width="10%">
        <b-switch :value="props.row.is_active" size="is-small"
          @input="toggleActive(props.row, $event)" />
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
        <div>
          <a href="#" @click.prevent="editSender(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(
            `Delete sender <b>${props.row.email}</b>?`,
            () => deleteSender(props.row),
          )" class="has-text-danger" aria-label="Delete">
            <b-icon icon="trash-can-outline" size="is-small" />
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <div class="has-text-centered has-text-grey py-4">No senders configured</div>
      </template>
    </b-table>

    <!-- New / Edit modal -->
    <b-modal v-model="isModalVisible" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Sender</p>
          <button type="button" class="delete" @click="isModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Email" message="The email address to send from">
            <b-input v-model="senderForm.email" type="email" required />
          </b-field>
          <b-field label="Display name" message="Shows as the sender name in recipient inbox">
            <b-input v-model="senderForm.name" />
          </b-field>
          <b-field label="Brand name" message="The brand this sender represents">
            <b-input v-model="senderForm.brand" />
          </b-field>
          <b-field label="Brand URL" message="The brand website URL">
            <b-input v-model="senderForm.brand_url" placeholder="https://..." />
          </b-field>
          <b-field label="Brand color" message="Used in email signatures and UI badges">
            <div class="is-flex is-align-items-center" style="gap: 0.75rem;">
              <input type="color" v-model="senderForm.brand_color" aria-label="Brand color picker"
                style="width: 48px; height: 36px; padding: 0; border: 1px solid #dbdbdb; border-radius: 4px; cursor: pointer;" />
              <b-input v-model="senderForm.brand_color" placeholder="#F2C94C"
                style="max-width: 120px;" />
            </div>
          </b-field>
          <b-field label="Active" v-if="isEditing">
            <b-switch v-model="senderForm.is_active" type="is-success" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isModalVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveSender" :loading="saving">
            {{ isEditing ? 'Save' : 'Add' }}
          </b-button>
        </footer>
      </div>
    </b-modal>
  </div>
</template>

<script>
import Vue from 'vue';

export default Vue.extend({
  props: {
    form: {
      type: Object, default: () => ({}),
    },
  },

  data() {
    return {
      senders: [],
      loading: false,
      saving: false,
      isModalVisible: false,
      isEditing: false,
      editingId: null,
      senderData: this.getEmptyForm(),
    };
  },

  computed: {
    domains() {
      const map = {};
      this.senders.forEach((s) => {
        const domain = s.email.split('@')[1];
        if (!domain) return;
        if (!map[domain]) {
          map[domain] = { domain, totalCount: 0, activeCount: 0 };
        }
        map[domain].totalCount += 1;
        if (s.is_active) {
          map[domain].activeCount += 1;
        }
      });
      return Object.values(map).sort((a, b) => a.domain.localeCompare(b.domain));
    },

    senderForm: {
      get() { return this.senderData; },
      set(v) { this.senderData = v; },
    },
  },

  methods: {
    getEmptyForm() {
      return {
        email: '',
        name: '',
        brand: '',
        brand_url: '',
        brand_color: '#F2C94C',
        is_active: true,
      };
    },

    getSenders() {
      this.loading = true;
      this.$api.getWarmingSenders().then((data) => {
        this.senders = data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewModal() {
      this.senderData = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.isModalVisible = true;
    },

    editSender(row) {
      this.senderData = {
        email: row.email,
        name: row.name,
        brand: row.brand,
        brand_url: row.brand_url,
        brand_color: row.brand_color || '#F2C94C',
        is_active: row.is_active,
      };
      this.isEditing = true;
      this.editingId = row.id;
      this.isModalVisible = true;
    },

    saveSender() {
      this.saving = true;
      const payload = {
        email: this.senderData.email,
        name: this.senderData.name,
        brand: this.senderData.brand,
        brand_url: this.senderData.brand_url,
        brand_color: this.senderData.brand_color,
      };
      if (this.isEditing) {
        payload.is_active = this.senderData.is_active;
      }

      const fn = this.isEditing
        ? this.$api.updateWarmingSender(this.editingId, payload)
        : this.$api.createWarmingSender(payload);

      fn.then(() => {
        this.$utils.toast(this.isEditing ? 'Sender updated' : 'Sender added');
        this.isModalVisible = false;
        this.saving = false;
        this.getSenders();
      }).catch(() => { this.saving = false; });
    },

    toggleActive(row, val) {
      this.$api.updateWarmingSender(row.id, {
        email: row.email,
        name: row.name,
        brand: row.brand,
        brand_url: row.brand_url,
        brand_color: row.brand_color,
        is_active: val,
      }).then(() => {
        this.getSenders();
      });
    },

    deleteSender(row) {
      this.$api.deleteWarmingSender(row.id).then(() => {
        this.$utils.toast('Sender deleted');
        this.getSenders();
      });
    },
  },

  mounted() {
    this.getSenders();
  },
});
</script>

<style scoped>
.color-swatch {
  display: inline-block;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid #dbdbdb;
}
</style>
