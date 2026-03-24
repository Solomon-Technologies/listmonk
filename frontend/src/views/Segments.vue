<template>
  <section class="segments">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          Segments
          <span v-if="!isNaN(segments.total)">({{ segments.total }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-button type="is-primary" icon-left="plus" @click="showNewForm" data-cy="btn-new">
          New
        </b-button>
      </div>
    </header>

    <b-table :data="segments.results" :loading="loading" hoverable paginated backend-pagination
      pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
      :per-page="segments.perPage" :total="segments.total" backend-sorting @sort="onSort">
      <b-table-column v-slot="props" field="name" label="Name" sortable :td-attrs="$utils.tdID">
        <a href="#" @click.prevent="editSegment(props.row)">
          <strong>{{ props.row.name }}</strong>
        </a>
        <p class="is-size-7 has-text-grey" v-if="props.row.description">{{ props.row.description }}</p>
        <b-taglist v-if="props.row.tags && props.row.tags.length">
          <b-tag class="is-small" v-for="t in props.row.tags" :key="t">{{ t }}</b-tag>
        </b-taglist>
      </b-table-column>

      <b-table-column v-slot="props" field="match_type" label="Match" width="10%">
        <b-tag>{{ props.row.matchType === 'all' ? 'ALL' : 'ANY' }}</b-tag>
      </b-table-column>

      <b-table-column v-slot="props" field="conditions" label="Conditions" width="10%">
        {{ props.row.conditions ? props.row.conditions.length : 0 }}
      </b-table-column>

      <b-table-column v-slot="props" field="subscriber_count" label="Subscribers" width="12%" sortable>
        <a href="#" @click.prevent="previewSubscribers(props.row)">
          {{ $utils.formatNumber(props.row.subscriberCount) }}
        </a>
      </b-table-column>

      <b-table-column v-slot="props" field="created_at" label="Created" width="15%" sortable>
        {{ $utils.niceDate(props.row.createdAt) }}
      </b-table-column>

      <b-table-column v-slot="props" cell-class="actions" width="10%" align="right">
        <div>
          <a href="#" @click.prevent="editSegment(props.row)" :aria-label="'Edit'">
            <b-tooltip label="Edit" type="is-dark">
              <b-icon icon="pencil-outline" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="refreshCount(props.row)" :aria-label="'Refresh count'">
            <b-tooltip label="Refresh count" type="is-dark">
              <b-icon icon="refresh" size="is-small" />
            </b-tooltip>
          </a>
          <a href="#" @click.prevent="$utils.confirm(null, () => deleteSegment(props.row))" :aria-label="'Delete'">
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

    <!-- Segment editor modal -->
    <b-modal v-model="isEditorVisible" :width="800" scroll="keep" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ isEditing ? 'Edit' : 'New' }} Segment</p>
          <button type="button" class="delete" @click="isEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-field label="Name">
            <b-input v-model="form.name" placeholder="Segment name" required />
          </b-field>
          <b-field label="Description">
            <b-input v-model="form.description" type="textarea" placeholder="Optional description" />
          </b-field>
          <b-field label="Match type">
            <b-select v-model="form.matchType">
              <option value="all">ALL conditions (AND)</option>
              <option value="any">ANY condition (OR)</option>
            </b-select>
          </b-field>
          <b-field label="Tags">
            <b-taginput v-model="form.tags" placeholder="Add tag" />
          </b-field>

          <hr />
          <h3 class="title is-6">Conditions</h3>
          <div v-for="(cond, i) in form.conditions" :key="i" class="columns is-vcentered condition-row">
            <div class="column is-3">
              <b-select v-model="cond.field" expanded placeholder="Field">
                <option v-for="f in conditionFields" :key="f" :value="f">{{ f }}</option>
              </b-select>
            </div>
            <div class="column is-3">
              <b-select v-model="cond.operator" expanded placeholder="Operator">
                <option v-for="op in conditionOperators" :key="op" :value="op">{{ op }}</option>
              </b-select>
            </div>
            <div class="column is-4">
              <b-input v-model="cond.value" expanded placeholder="Value" />
            </div>
            <div class="column is-2">
              <a href="#" @click.prevent="removeCondition(i)" aria-label="Remove condition">
                <b-icon icon="close-circle-outline" size="is-small" />
              </a>
            </div>
          </div>
          <b-button size="is-small" icon-left="plus" @click="addCondition">
            Add condition
          </b-button>

          <hr />
          <div class="has-text-right">
            <b-button size="is-small" @click="previewCount" :loading="countLoading">
              Preview count: {{ previewCountValue !== null ? previewCountValue : '?' }}
            </b-button>
          </div>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveSegment" :loading="saving">Save</b-button>
        </footer>
      </div>
    </b-modal>

    <!-- Subscribers preview modal -->
    <b-modal v-model="isPreviewVisible" :width="700" scroll="keep" has-modal-card>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">Subscribers in "{{ previewSegment ? previewSegment.name : '' }}"</p>
          <button type="button" class="delete" @click="isPreviewVisible = false" />
        </header>
        <section class="modal-card-body">
          <b-table :data="previewSubs" :loading="previewLoading" hoverable>
            <b-table-column v-slot="props" field="email" label="Email">
              <router-link :to="{ name: 'subscriber', params: { id: props.row.id } }">
                {{ props.row.email }}
              </router-link>
            </b-table-column>
            <b-table-column v-slot="props" field="name" label="Name">
              {{ props.row.name }}
            </b-table-column>
            <b-table-column v-slot="props" field="status" label="Status">
              <b-tag :class="props.row.status">{{ props.row.status }}</b-tag>
            </b-table-column>
          </b-table>
        </section>
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
      segments: { results: [], total: 0, perPage: 20 },
      loading: false,
      saving: false,
      countLoading: false,
      previewCountValue: null,

      queryParams: {
        page: 1,
        orderBy: 'created_at',
        order: 'desc',
      },

      isEditorVisible: false,
      isEditing: false,
      editingId: null,
      form: this.getEmptyForm(),

      conditionFields: [
        'email', 'name', 'status', 'created_at', 'lists', 'score',
        'attribs.city', 'attribs.country', 'attribs.company',
      ],
      conditionOperators: [
        'eq', 'neq', 'contains', 'not_contains', 'gt', 'lt', 'gte', 'lte',
        'starts_with', 'ends_with', 'is_set', 'is_not_set', 'in_list', 'not_in_list',
      ],

      isPreviewVisible: false,
      previewSegment: null,
      previewSubs: [],
      previewLoading: false,
    };
  },

  methods: {
    getEmptyForm() {
      return {
        name: '',
        description: '',
        matchType: 'all',
        tags: [],
        conditions: [],
      };
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getSegments();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getSegments();
    },

    getSegments() {
      this.loading = true;
      this.$api.getSegments({
        page: this.queryParams.page,
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      }).then((data) => {
        this.segments = data;
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    showNewForm() {
      this.form = this.getEmptyForm();
      this.isEditing = false;
      this.editingId = null;
      this.previewCountValue = null;
      this.isEditorVisible = true;
    },

    editSegment(s) {
      this.form = {
        name: s.name,
        description: s.description || '',
        matchType: s.matchType || 'all',
        tags: s.tags || [],
        conditions: (s.conditions || []).map((c) => ({ ...c })),
      };
      this.isEditing = true;
      this.editingId = s.id;
      this.previewCountValue = s.subscriberCount;
      this.isEditorVisible = true;
    },

    addCondition() {
      this.form.conditions.push({ field: 'email', operator: 'contains', value: '' });
    },

    removeCondition(i) {
      this.form.conditions.splice(i, 1);
    },

    saveSegment() {
      this.saving = true;
      const data = {
        name: this.form.name,
        description: this.form.description,
        match_type: this.form.matchType,
        tags: this.form.tags,
        conditions: this.form.conditions,
      };

      const fn = this.isEditing
        ? this.$api.updateSegment(this.editingId, data)
        : this.$api.createSegment(data);

      fn.then(() => {
        this.isEditorVisible = false;
        this.getSegments();
        this.$utils.toast(this.isEditing ? 'Segment updated' : 'Segment created');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    deleteSegment(s) {
      this.$api.deleteSegment(s.id).then(() => {
        this.getSegments();
        this.$utils.toast(`Deleted "${s.name}"`);
      });
    },

    refreshCount(s) {
      const seg = s;
      this.$api.getSegmentCount(seg.id).then((data) => {
        seg.subscriberCount = data.count;
        this.$utils.toast(`${seg.name}: ${data.count} subscribers`);
      });
    },

    previewCount() {
      if (!this.editingId) {
        this.previewCountValue = '(save first)';
        return;
      }
      this.countLoading = true;
      this.$api.getSegmentCount(this.editingId).then((data) => {
        this.previewCountValue = data.count;
        this.countLoading = false;
      }).catch(() => { this.countLoading = false; });
    },

    previewSubscribers(s) {
      this.previewSegment = s;
      this.previewLoading = true;
      this.previewSubs = [];
      this.isPreviewVisible = true;
      this.$api.getSegmentSubscribers(s.id, { per_page: 50 }).then((data) => {
        this.previewSubs = data.results || [];
        this.previewLoading = false;
      }).catch(() => { this.previewLoading = false; });
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getSegments);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.getSegments);
  },

  mounted() {
    this.getSegments();
  },
});
</script>

<style scoped>
.condition-row {
  margin-bottom: 0.25rem;
}
</style>
