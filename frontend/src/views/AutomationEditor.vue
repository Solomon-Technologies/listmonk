<template>
  <section class="automation-editor">
    <header class="columns page-header">
      <div class="column is-8">
        <h1 class="title is-4">
          <router-link :to="{ name: 'automations' }">Automations</router-link>
          <span v-if="automation.name"> / {{ automation.name }}</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <b-tag :class="automation.status" size="is-medium">{{ automation.status }}</b-tag>
        <b-button v-if="automation.status === 'draft'" type="is-success" size="is-small"
          @click="changeStatus('active')">
          Activate
        </b-button>
        <b-button v-if="automation.status === 'active'" type="is-warning" size="is-small"
          @click="changeStatus('paused')">
          Pause
        </b-button>
        <b-button v-if="automation.status === 'paused'" type="is-success" size="is-small"
          @click="changeStatus('active')">
          Resume
        </b-button>
      </div>
    </header>

    <b-loading :is-full-page="false" v-model="loading" />

    <div class="columns" v-if="!loading">
      <!-- Settings -->
      <div class="column is-3">
        <div class="box">
          <h3 class="title is-6">Settings</h3>
          <b-field label="Name">
            <b-input v-model="automation.name" />
          </b-field>
          <b-field label="Description">
            <b-input v-model="automation.description" type="textarea" />
          </b-field>
          <hr />
          <div class="is-size-7">
            <p>Entered: {{ automation.totalEntered || 0 }}</p>
            <p>Completed: {{ automation.totalCompleted || 0 }}</p>
          </div>
          <hr />
          <b-button type="is-primary" expanded size="is-small" @click="saveAutomation" :loading="saving">
            Save
          </b-button>
        </div>

        <!-- Node palette -->
        <div class="box">
          <h3 class="title is-6">Add Node</h3>
          <div class="buttons are-small">
            <b-button v-for="nt in nodeTypes" :key="nt.type" @click="addNode(nt.type)" size="is-small"
              :icon-left="nt.icon" expanded class="mb-1">
              {{ nt.label }}
            </b-button>
          </div>
        </div>
      </div>

      <!-- Canvas -->
      <div class="column is-9">
        <div class="box automation-canvas" ref="canvas">
          <div v-if="nodes.length === 0" class="has-text-centered has-text-grey py-6">
            Add nodes from the left panel to build your automation flow.
          </div>

          <!-- Node list (simplified visual) -->
          <div v-for="(node, i) in nodes" :key="node.id" class="automation-node"
            :class="'node-' + node.nodeType">
            <div class="columns is-vcentered">
              <div class="column is-1 has-text-centered">
                <b-icon :icon="getNodeIcon(node.nodeType)" />
              </div>
              <div class="column is-4">
                <strong>{{ getNodeLabel(node.nodeType) }}</strong>
                <p class="is-size-7 has-text-grey">{{ getNodeSummary(node) }}</p>
              </div>
              <div class="column is-5">
                <a href="#" @click.prevent="editNode(node)" class="is-size-7">
                  Configure
                </a>
              </div>
              <div class="column is-2 has-text-right">
                <a href="#" @click.prevent="$utils.confirm(null, () => deleteNode(node))" aria-label="Delete node">
                  <b-icon icon="close-circle-outline" size="is-small" />
                </a>
              </div>
            </div>
            <div v-if="i < nodes.length - 1" class="node-connector has-text-centered">
              <b-icon icon="arrow-down" size="is-small" class="has-text-grey-light" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Node config modal -->
    <b-modal v-model="isNodeEditorVisible" :width="600" has-modal-card>
      <div class="modal-card" v-if="editingNode">
        <header class="modal-card-head">
          <p class="modal-card-title">Configure {{ getNodeLabel(editingNode.nodeType) }}</p>
          <button type="button" class="delete" @click="isNodeEditorVisible = false" />
        </header>
        <section class="modal-card-body">
          <!-- Trigger config -->
          <template v-if="editingNode.nodeType === 'trigger'">
            <b-field label="Trigger event">
              <b-select v-model="nodeConfig.event" expanded>
                <option value="subscriber.created">Subscriber created</option>
                <option value="subscriber.optin">Subscriber opted in</option>
                <option value="tag.added">Tag added</option>
                <option value="list.subscribed">List subscription</option>
              </b-select>
            </b-field>
          </template>

          <!-- Email action -->
          <template v-if="editingNode.nodeType === 'action_email'">
            <b-field label="Subject">
              <b-input v-model="nodeConfig.subject" />
            </b-field>
            <b-field label="Body">
              <b-input v-model="nodeConfig.body" type="textarea" rows="6" />
            </b-field>
          </template>

          <!-- Delay -->
          <template v-if="editingNode.nodeType === 'delay'">
            <div class="columns">
              <div class="column">
                <b-field label="Delay value">
                  <b-numberinput v-model="nodeConfig.delay_value" :min="1" />
                </b-field>
              </div>
              <div class="column">
                <b-field label="Unit">
                  <b-select v-model="nodeConfig.delay_unit" expanded>
                    <option value="minutes">Minutes</option>
                    <option value="hours">Hours</option>
                    <option value="days">Days</option>
                    <option value="weeks">Weeks</option>
                  </b-select>
                </b-field>
              </div>
            </div>
          </template>

          <!-- Condition -->
          <template v-if="editingNode.nodeType === 'condition'">
            <b-field label="Condition field">
              <b-input v-model="nodeConfig.field" placeholder="e.g. attribs.plan" />
            </b-field>
            <b-field label="Operator">
              <b-select v-model="nodeConfig.operator" expanded>
                <option value="eq">Equals</option>
                <option value="neq">Not equals</option>
                <option value="contains">Contains</option>
                <option value="gt">Greater than</option>
                <option value="lt">Less than</option>
              </b-select>
            </b-field>
            <b-field label="Value">
              <b-input v-model="nodeConfig.value" />
            </b-field>
          </template>

          <!-- Tag action -->
          <template v-if="editingNode.nodeType === 'action_tag'">
            <b-field label="Tag name">
              <b-input v-model="nodeConfig.tag" placeholder="Tag to add" />
            </b-field>
            <b-field label="Action">
              <b-select v-model="nodeConfig.action" expanded>
                <option value="add">Add tag</option>
                <option value="remove">Remove tag</option>
              </b-select>
            </b-field>
          </template>

          <!-- List action -->
          <template v-if="editingNode.nodeType === 'action_list'">
            <b-field label="List ID">
              <b-numberinput v-model="nodeConfig.list_id" :min="1" />
            </b-field>
            <b-field label="Action">
              <b-select v-model="nodeConfig.action" expanded>
                <option value="subscribe">Subscribe</option>
                <option value="unsubscribe">Unsubscribe</option>
              </b-select>
            </b-field>
          </template>

          <!-- Webhook action -->
          <template v-if="editingNode.nodeType === 'action_webhook'">
            <b-field label="Webhook URL">
              <b-input v-model="nodeConfig.url" type="url" />
            </b-field>
          </template>
        </section>
        <footer class="modal-card-foot has-text-right">
          <b-button @click="isNodeEditorVisible = false">Cancel</b-button>
          <b-button type="is-primary" @click="saveNodeConfig" :loading="savingNode">Save</b-button>
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
      automation: {},
      nodes: [],
      edges: [],
      loading: true,
      saving: false,
      savingNode: false,

      nodeTypes: [
        { type: 'trigger', label: 'Trigger', icon: 'flash-outline' },
        { type: 'delay', label: 'Delay', icon: 'clock-outline' },
        { type: 'action_email', label: 'Send Email', icon: 'email-outline' },
        { type: 'condition', label: 'Condition', icon: 'help-circle-outline' },
        { type: 'action_tag', label: 'Tag', icon: 'tag-outline' },
        { type: 'action_list', label: 'List Action', icon: 'format-list-bulleted-square' },
        { type: 'action_webhook', label: 'Webhook', icon: 'webhook' },
      ],

      isNodeEditorVisible: false,
      editingNode: null,
      nodeConfig: {},
    };
  },

  methods: {
    getNodeIcon(type) {
      const t = this.nodeTypes.find((n) => n.type === type);
      return t ? t.icon : 'help-circle-outline';
    },

    getNodeLabel(type) {
      const t = this.nodeTypes.find((n) => n.type === type);
      return t ? t.label : type;
    },

    getNodeSummary(node) {
      const c = node.config || {};
      switch (node.nodeType) {
        case 'trigger': return c.event || 'Not configured';
        case 'delay': return `Wait ${c.delay_value || '?'} ${c.delay_unit || 'days'}`;
        case 'action_email': return c.subject || 'No subject';
        case 'condition': return `${c.field || '?'} ${c.operator || '?'} ${c.value || '?'}`;
        case 'action_tag': return `${c.action || 'add'} tag "${c.tag || '?'}"`;
        case 'action_list': return `${c.action || 'subscribe'} list #${c.list_id || '?'}`;
        case 'action_webhook': return c.url || 'No URL';
        default: return '';
      }
    },

    getAutomation() {
      const id = parseInt(this.$route.params.id, 10);
      this.loading = true;
      this.$api.getAutomation(id).then((data) => {
        this.automation = data;
        this.nodes = data.nodes || [];
        this.edges = data.edges || [];
        this.loading = false;
      }).catch(() => { this.loading = false; });
    },

    saveAutomation() {
      this.saving = true;
      this.$api.updateAutomation(this.automation.id, {
        name: this.automation.name,
        description: this.automation.description,
        status: this.automation.status,
      }).then(() => {
        this.$utils.toast('Automation saved');
        this.saving = false;
      }).catch(() => { this.saving = false; });
    },

    changeStatus(status) {
      this.$api.updateAutomation(this.automation.id, {
        ...this.automation,
        status,
      }).then(() => {
        this.automation.status = status;
        this.$utils.toast(`Status changed to ${status}`);
      });
    },

    addNode(nodeType) {
      const id = parseInt(this.$route.params.id, 10);
      this.$api.createAutomationNode(id, {
        node_type: nodeType,
        config: {},
        position_x: 0,
        position_y: this.nodes.length * 120,
      }).then(() => {
        this.getAutomation();
        this.$utils.toast(`${this.getNodeLabel(nodeType)} added`);
      });
    },

    editNode(node) {
      this.editingNode = node;
      this.nodeConfig = { ...(node.config || {}) };
      this.isNodeEditorVisible = true;
    },

    saveNodeConfig() {
      this.savingNode = true;
      const id = parseInt(this.$route.params.id, 10);
      this.$api.updateAutomationNode(id, this.editingNode.id, {
        ...this.editingNode,
        config: this.nodeConfig,
      }).then(() => {
        this.isNodeEditorVisible = false;
        this.getAutomation();
        this.$utils.toast('Node updated');
        this.savingNode = false;
      }).catch(() => { this.savingNode = false; });
    },

    deleteNode(node) {
      const id = parseInt(this.$route.params.id, 10);
      this.$api.deleteAutomationNode(id, node.id).then(() => {
        this.getAutomation();
        this.$utils.toast('Node removed');
      });
    },
  },

  created() {
    this.$root.$on('page.refresh', () => {
      this.getAutomation();
    });
  },

  destroyed() {
    this.$root.$off('page.refresh');
  },

  mounted() {
    this.getAutomation();
  },
});
</script>

<style scoped>
.automation-canvas {
  min-height: 500px;
}
.automation-node {
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  padding: 0.75rem;
  margin-bottom: 0.25rem;
  transition: border-color 0.2s;
}
.automation-node:hover {
  border-color: #3273dc;
}
.node-trigger { border-left: 4px solid #48c774; }
.node-delay { border-left: 4px solid #ffdd57; }
.node-action_email { border-left: 4px solid #3273dc; }
.node-condition { border-left: 4px solid #ff3860; }
.node-action_tag { border-left: 4px solid #00d1b2; }
.node-action_list { border-left: 4px solid #7957d5; }
.node-action_webhook { border-left: 4px solid #f14668; }
.node-connector {
  padding: 0.15rem 0;
}
</style>
