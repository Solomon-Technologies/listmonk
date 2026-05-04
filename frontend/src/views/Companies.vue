<template>
  <section class="companies">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Companies
          <span v-if="companies && companies.length">({{ companies.length }})</span>
        </h1>
        <p class="subtitle is-size-7 has-text-grey">
          Tenants in this Listmonk instance. Each tenant has its own lists,
          subscribers, campaigns, templates, warming, and roles. Users are
          scoped to one tenant.
        </p>
      </div>
      <div class="column has-text-right">
        <b-field v-if="$can('settings:manage')" expanded>
          <b-button expanded type="is-primary" icon-left="plus"
                    class="btn-new" @click="showNewForm">
            New tenant
          </b-button>
        </b-field>
      </div>
    </header>

    <b-table :data="rows" :loading="loading.companies" hoverable
             default-sort="id">
      <b-table-column v-slot="props" field="id" label="ID" sortable width="60">
        {{ props.row.id }}
      </b-table-column>

      <b-table-column v-slot="props" field="name" label="Name" sortable>
        <a v-if="$can('settings:manage')" :href="`/companies/${props.row.id}`"
           @click.prevent="showEditForm(props.row)">
          {{ props.row.name }}
        </a>
        <span v-else>{{ props.row.name }}</span>
      </b-table-column>

      <b-table-column v-slot="props" field="slug" label="Slug" sortable>
        <code>{{ props.row.slug }}</code>
      </b-table-column>

      <b-table-column v-slot="props" label="Tenant data">
        <div class="is-size-7">
          <span class="mr-2">
            <b-icon icon="account-outline" size="is-small" />
            {{ props.row.users || 0 }} users
          </span>
          <span class="mr-2">
            <b-icon icon="format-list-bulleted" size="is-small" />
            {{ props.row.lists || 0 }} lists
          </span>
          <span class="mr-2">
            <b-icon icon="account-multiple" size="is-small" />
            {{ (props.row.subscribers || 0).toLocaleString() }} subs
          </span>
          <span class="mr-2">
            <b-icon icon="email-multiple-outline" size="is-small" />
            {{ props.row.campaigns || 0 }} campaigns
          </span>
          <span class="mr-2">
            <b-icon icon="file-document-outline" size="is-small" />
            {{ props.row.templates || 0 }} templates
          </span>
          <span class="mr-2">
            <b-icon icon="fire" size="is-small" />
            {{ props.row.warming_campaigns || 0 }} warming
          </span>
          <span class="mr-2">
            <b-icon icon="shield-account-outline" size="is-small" />
            {{ props.row.roles || 0 }} roles
          </span>
        </div>
      </b-table-column>

      <b-table-column v-slot="props" label="Created">
        <span class="is-size-7" v-if="props.row.created_at">
          {{ $utils.niceDate(props.row.created_at) }}
        </span>
      </b-table-column>

      <b-table-column v-slot="props" v-if="$can('settings:manage')" label="">
        <b-button @click="showEditForm(props.row)" icon-left="pencil"
                  type="is-text" size="is-small" />
        <b-button @click="confirmDelete(props.row)" icon-left="trash-can-outline"
                  type="is-text" size="is-small"
                  :disabled="(props.row.users || 0) > 0
                             || (props.row.lists || 0) > 0
                             || (props.row.subscribers || 0) > 0
                             || (props.row.campaigns || 0) > 0" />
      </b-table-column>
    </b-table>

    <b-modal :active.sync="modal.active" scroll="keep" :width="600"
             :on-cancel="closeModal">
      <form @submit.prevent="onSubmit">
        <div class="modal-card content company-modal" style="width: auto">
          <header class="modal-card-head">
            <h4>{{ modal.editing ? 'Edit tenant' : 'New tenant' }}</h4>
          </header>
          <section class="modal-card-body modal-fullheight">
            <b-field label="Name" label-position="on-border">
              <b-input v-model="form.name" maxlength="200" required
                       placeholder="Acme Co" ref="focus" />
            </b-field>

            <b-field label="Slug" label-position="on-border"
                     message="Lowercase letters, digits, hyphens. Used in URLs and as a stable identifier.">
              <b-input v-model="form.slug" maxlength="80" required
                       placeholder="acme-co" />
            </b-field>

            <p class="has-text-grey is-size-7 mt-4" v-if="!modal.editing">
              After creating the tenant, default Super Admin and
              Operational Admin roles are NOT auto-created (only the v7.17.0
              install seeded those for Solomon &amp; Rule27). Create roles
              for this new tenant under Settings → Users → Roles, then assign
              users to them.
            </p>
          </section>
          <footer class="modal-card-foot has-text-right">
            <b-button @click="closeModal">Cancel</b-button>
            <b-button native-type="submit" type="is-primary"
                      :loading="loading.companies">
              Save
            </b-button>
          </footer>
        </div>
      </form>
    </b-modal>
  </section>
</template>

<script>
import Vue from 'vue';
import { mapState } from 'vuex';

export default Vue.extend({
  name: 'Companies',

  data() {
    return {
      // The /api/companies/stats response merged onto the row list.
      rows: [],
      modal: {
        active: false,
        editing: false,
      },
      form: { id: 0, name: '', slug: '' },
    };
  },

  computed: {
    ...mapState(['loading', 'companies']),
  },

  methods: {
    fetchRows() {
      // /api/companies/stats requires settings:manage. If the user only
      // has users:get they fall back to the lighter /api/companies list.
      if (this.$can('settings:manage')) {
        this.$api.getCompanyStats().then((d) => { this.rows = d; });
      } else {
        this.$api.getCompanies().then((d) => { this.rows = d; });
      }
    },

    showNewForm() {
      this.form = { id: 0, name: '', slug: '' };
      this.modal = { active: true, editing: false };
      this.$nextTick(() => {
        if (this.$refs.focus) this.$refs.focus.focus();
      });
    },

    showEditForm(row) {
      this.form = { id: row.id, name: row.name, slug: row.slug };
      this.modal = { active: true, editing: true };
    },

    closeModal() {
      this.modal = { active: false, editing: false };
    },

    onSubmit() {
      const payload = {
        id: this.form.id,
        name: this.form.name.trim(),
        slug: this.form.slug.trim().toLowerCase(),
      };
      const op = this.modal.editing
        ? this.$api.updateCompany(payload)
        : this.$api.createCompany(payload);

      op.then(() => {
        this.$utils.toast(this.modal.editing ? 'Tenant updated' : 'Tenant created');
        this.closeModal();
        this.fetchRows();
      });
    },

    confirmDelete(row) {
      this.$utils.confirm(
        `Delete tenant "${row.name}"? This requires zero remaining users, lists, subscribers, campaigns, templates, warming records, or roles.`,
        () => {
          this.$api.deleteCompany(row.id).then(() => {
            this.$utils.toast(`Tenant "${row.name}" deleted`);
            this.fetchRows();
          });
        },
      );
    },
  },

  mounted() {
    this.fetchRows();
  },
});
</script>
