<template>
  <section class="warming-recipients">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Warming Recipients
          <span v-if="!isNaN(addresses.length)">({{ addresses.length }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewModal">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="addresses" :loading="loading" hoverable>
      <b-table-column v-slot="props" field="email" label="Email">
        {{ props.row.email }}
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name">
        {{ props.row.name }}
      </b-table-column>

      <b-table-column v-slot="props" field="is_active" label="Active" width="10%">
        <b-switch :value="props.row.is_active" size="is-small"
          @input="toggleActive(props.row, $event)" />
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="15%">
        {{ $utils.niceDate(props.row.created_at) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
        <div>
          <a href="#" @click.prevent="editAddress(props.row)" aria-label="Edit">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(
            `Delete warming recipient <b>${props.row.email}</b>?`,
            () => deleteAddress(props.row),
          )" class="has-text-danger" aria-label="Delete">
            <b-icon icon="trash-can-outline" size="is-small" />
          </a>
        </div>
      </b-table-column>

      <template #empty v-if="!loading">
        <empty-placeholder />
      </template>
    </b-table>

    <!-- New / Edit modal -->
    <b-modal v-model="isModalVisible" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Warming Recipient</p>
          <button type="button" class="delete" @click="isModalVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Email">
            <b-input v-model="form.email" type="email" required />
          </b-field>
          <b-field label="Name">
            <b-input v-model="form.name" />
          </b-field>
          <b-field label="Active" v-if="isEditing">
            <b-switch v-model="form.is_active" />
          </b-field>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isModalVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveAddress" :loading="saving">
            {{ isEditing ? 'Save' : 'Add' }}
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
      addresses: [],
      loading: false,
      saving: false,
      isModalVisible: false,
      isEditing: false,
      editingId: null,
      form: { email: '', name: '', is_active: true },
    };
  },

  methods: {
    getAddresses() {
      this.loading = true;
      this.$api.getWarmingAddresses().then((data) => {
        this.addresses = data || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewModal() {
      this.form = { email: '', name: '', is_active: true };
      this.isEditing = false;
      this.editingId = null;
      this.isModalVisible = true;
    },

    editAddress(row) {
      this.form = {
        email: row.email,
        name: row.name,
        is_active: row.is_active,
      };
      this.isEditing = true;
      this.editingId = row.id;
      this.isModalVisible = true;
    },

    saveAddress() {
      this.saving = true;

      const fn = this.isEditing
        ? this.$api.updateWarmingAddress(this.editingId, {
          email: this.form.email,
          name: this.form.name,
          is_active: this.form.is_active,
        })
        : this.$api.createWarmingAddress({
          email: this.form.email,
          name: this.form.name,
        });

      fn.then(() => {
        this.$utils.toast(this.isEditing ? 'Recipient updated' : 'Recipient added');
        this.isModalVisible = false;
        this.saving = false;
        this.getAddresses();
      }).catch(() => { this.saving = false; });
    },

    toggleActive(row, val) {
      this.$api.updateWarmingAddress(row.id, {
        email: row.email,
        name: row.name,
        is_active: val,
      }).then(() => {
        this.getAddresses();
      });
    },

    deleteAddress(row) {
      this.$api.deleteWarmingAddress(row.id).then(() => {
        this.$utils.toast('Recipient deleted');
        this.getAddresses();
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getAddresses);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getAddresses);
  },

  mounted() {
    this.getAddresses();
  },
});
</script>
