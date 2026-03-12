<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items != null">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>
        {{ $t('dashboard2') }}
      </v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <div class="history-shell px-4 pb-6">
      <v-row dense class="mt-4">
        <v-col
          v-for="card in summaryCards"
          :key="card.title"
          cols="12"
          sm="6"
          xl="3"
        >
          <v-card class="history-kpi" outlined>
            <div class="history-kpi__label">{{ card.title }}</div>
            <div class="history-kpi__value">{{ card.value }}</div>
          </v-card>
        </v-col>
      </v-row>

      <v-card class="history-panel mt-4" outlined>
        <div class="history-panel__header">
          <div>
            <div class="history-panel__title">{{ $t('history') }}</div>
            <div class="history-panel__subtitle">{{ $t('overview_recent_runs_subtitle') }}</div>
          </div>

          <v-chip small outlined class="history-panel__chip">
            {{ formatNumber(items.length) }}
          </v-chip>
        </div>

        <v-data-table
          :headers="headers"
          :items="items"
          :footer-props="{ itemsPerPageOptions: [20] }"
          class="mt-4 HistoryTable"
        >
          <template v-slot:item.tpl_alias="{ item }">
            <div class="d-flex align-center">
              <v-icon
                class="mr-3"
                small
              >
                {{ getAppIcon(item.tpl_app) }}
              </v-icon>

              <TaskLink
                :task-id="item.id"
                :label="'#' + item.id"
              />

              <v-icon small class="ml-1 mr-1">mdi-arrow-left</v-icon>

              <router-link :to="
                '/project/' + item.project_id +
                '/templates/' + item.template_id"
              >{{ item.tpl_alias }}
              </router-link>
            </div>

            <div style="font-size: 14px;" class="ml-7">
                <span v-if="item.message">
                  <v-icon x-small>mdi-message-outline</v-icon> {{ item.message }}
                </span>
              <span v-else-if="item.commit_hash">
                  <v-icon x-small>mdi-source-fork</v-icon> {{ item.commit_message }}
                </span>
            </div>
          </template>

          <template v-slot:item.version="{ item }">
            <TaskLink
              :disabled="item.tpl_type === 'build'"
              class="ml-2"
              v-if="item.tpl_type !== ''"
              :status="item.status"

              :task-id="item.tpl_type === 'build'
                  ? item.id
                  : (item.build_task || {}).id"

              :label="item.tpl_type === 'build'
                  ? item.version
                  : (item.build_task || {}).version"

              :tooltip="item.tpl_type === 'build'
                  ? item.message
                  : (item.build_task || {}).message"
            />
            <div class="ml-2" v-else>&mdash;</div>
          </template>

          <template v-slot:item.status="{ item }">
            <TaskStatus :status="item.status"/>
          </template>

          <template v-slot:item.start="{ item }">
            {{ item.start | formatDate }}
          </template>

          <template v-slot:item.end="{ item }">
            {{ [item.start, item.end] | formatMilliseconds }}
          </template>
        </v-data-table>
      </v-card>
    </div>
  </div>
</template>

<style lang="scss">
.history-shell {
  margin: auto;
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 56px);
}

.history-kpi,
.history-panel {
  border-radius: 22px !important;
}

.history-kpi {
  height: 100%;
  padding: 20px;
}

.history-kpi__label {
  color: #607d8b;
  font-size: 13px;
}

.history-kpi__value {
  color: #102027;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: -0.04em;
  margin-top: 8px;
}

.history-panel {
  padding: 22px;
}

.history-panel__header {
  align-items: center;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.history-panel__title {
  color: #102027;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.history-panel__subtitle {
  color: #607d8b;
  font-size: 13px;
  margin-top: 4px;
}

.history-panel__chip {
  font-weight: 700;
}

.HistoryTable td {
  height: 60px !important;
}

.theme--dark {
  .history-kpi__value,
  .history-panel__title {
    color: #eceff1;
  }

  .history-kpi__label,
  .history-panel__subtitle {
    color: #90a4ae;
  }
}

@media (max-width: 1264px) {
  .history-shell {
    max-width: calc(100vw - 32px);
  }
}

@media (max-width: 960px) {
  .history-shell {
    max-width: 100%;
  }

  .history-panel__header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>

<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import EventBus from '@/event-bus';
import TaskStatus from '@/components/TaskStatus.vue';
import TaskLink from '@/components/TaskLink.vue';
import socket from '@/socket';
import { TEMPLATE_TYPE_ICONS } from '@/lib/constants';
import AppsMixin from '@/components/AppsMixin';
import DashboardMenu from '@/components/DashboardMenu.vue';

const ACTIVE_STATUSES = ['waiting', 'starting', 'running', 'stopping', 'waiting_confirmation', 'confirmed'];

export default {
  mixins: [ItemListPageBase, AppsMixin],

  data() {
    return { TEMPLATE_TYPE_ICONS };
  },

  components: { DashboardMenu, TaskStatus, TaskLink },

  computed: {
    finishedRuns() {
      return (this.items || [])
        .filter((item) => ['success', 'error', 'stopped'].includes(item.status))
        .length;
    },

    successRate() {
      if (!this.finishedRuns) {
        return 0;
      }

      return (this.items.filter((item) => item.status === 'success').length / this.finishedRuns) * 100;
    },

    summaryCards() {
      const items = this.items || [];

      return [{
        title: this.$t('overview_total_runs'),
        value: this.formatNumber(items.length),
      }, {
        title: this.$t('overview_success_rate'),
        value: `${this.successRate.toFixed(1)}%`,
      }, {
        title: this.$t('overview_failed_runs'),
        value: this.formatNumber(items.filter((item) => item.status === 'error').length),
      }, {
        title: this.$t('overview_active_runs'),
        value: this.formatNumber(
          items.filter((item) => ACTIVE_STATUSES.includes(item.status)).length,
        ),
      }];
    },
  },

  watch: {
    async projectId() {
      await this.loadItems();
    },
  },

  created() {
    socket.addListener((data) => this.onWebsocketDataReceived(data));
  },

  methods: {
    showTaskLog(taskId) {
      EventBus.$emit('i-show-task', {
        taskId,
      });
    },

    async onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      if (!this.items.some((item) => item.id === data.task_id)) {
        await this.loadItems();
      }

      const task = this.items.find((item) => item.id === data.task_id);

      if (task) {
        Object.assign(task, {
          ...data,
          type: undefined,
        });
      }
    },

    formatNumber(value, maximumFractionDigits = 0) {
      return new Intl.NumberFormat(undefined, {
        minimumFractionDigits: maximumFractionDigits > 0 ? 1 : 0,
        maximumFractionDigits,
      }).format(value || 0);
    },

    getHeaders() {
      return [
        {
          text: this.$i18n.t('task2'),
          value: 'tpl_alias',
          sortable: false,
        },
        {
          text: this.$i18n.t('version'),
          value: 'version',
          sortable: false,
        },
        {
          text: this.$i18n.t('status'),
          value: 'status',
          sortable: false,
        },
        {
          text: this.$i18n.t('user'),
          value: 'user_name',
          sortable: false,
        },
        {
          text: this.$i18n.t('start'),
          value: 'start',
          sortable: false,
        },
        {
          text: this.$i18n.t('duration'),
          value: 'end',
          sortable: false,
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/tasks/last`;
    },
  },
};
</script>
