<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <div class="overview-shell px-4 pb-6">
      <v-progress-linear
        v-if="isLoading"
        indeterminate
        color="primary"
        class="overview-loading"
      />

      <div class="overview-toolbar mt-4">
        <div class="overview-toolbar__actions">
          <v-btn-toggle
            v-model="dateRange"
            mandatory
            tile
            class="overview-range-toggle"
          >
            <v-btn
              v-for="range in dateRanges"
              :key="range.value"
              :value="range.value"
              small
              depressed
            >
              {{ range.label }}
            </v-btn>
          </v-btn-toggle>

          <v-select
            v-model="selectedUser"
            :items="users"
            dense
            outlined
            hide-details
            item-text="text"
            item-value="value"
            class="overview-user-filter"
            prepend-inner-icon="mdi-account-outline"
          />
        </div>
      </div>

      <v-row dense class="mt-4">
        <v-col
          v-for="card in summaryCards"
          :key="card.title"
          cols="12"
          sm="6"
          xl="3"
        >
          <v-card class="overview-metric-card" outlined>
            <div class="overview-metric-card__top">
              <div
                class="overview-metric-card__icon"
                :class="`overview-metric-card__icon--${card.tone}`"
              >
                <v-icon small>{{ card.icon }}</v-icon>
              </div>

              <div
                v-if="card.delta.text"
                class="overview-delta-chip"
                :class="`overview-delta-chip--${card.delta.tone}`"
              >
                {{ card.delta.text }}
              </div>
            </div>

            <div class="overview-metric-card__label">{{ card.title }}</div>
            <div class="overview-metric-card__value">{{ card.value }}</div>
            <div class="overview-metric-card__caption">{{ $t('overview_vs_previous_period') }}</div>
          </v-card>
        </v-col>
      </v-row>

      <v-row dense class="mt-1">
        <v-col cols="12" lg="8">
          <v-card class="overview-panel" outlined>
            <div class="overview-panel__header">
              <div>
                <div class="overview-panel__title">{{ $t('overview_execution_trend') }}</div>
                <div class="overview-panel__subtitle">
                  {{ $t('overview_execution_trend_subtitle', { days: rangeDays }) }}
                </div>
              </div>

              <div class="overview-panel__summary">
                <div class="overview-panel__summary-value">{{ formattedTotalRuns }}</div>
                <div class="overview-panel__summary-label">{{ $t('overview_total_runs') }}</div>
              </div>
            </div>

            <div class="overview-chart-wrapper">
              <OverviewTrendChart
                :source-data="normalizedStats"
                :dark="$vuetify.theme.dark"
              />
            </div>
          </v-card>
        </v-col>

        <v-col cols="12" lg="4">
          <v-card class="overview-panel" outlined>
            <div class="overview-panel__title">{{ $t('overview_status_breakdown') }}</div>
            <div class="overview-panel__subtitle">
              {{ $t('overview_status_breakdown_subtitle') }}
            </div>

            <div class="overview-status-list">
              <div
                v-for="status in statusBreakdown"
                :key="status.key"
                class="overview-status-row"
              >
                <div class="overview-status-row__header">
                  <span>{{ status.label }}</span>
                  <span>{{ status.countLabel }}</span>
                </div>

                <v-progress-linear
                  :value="status.percentage"
                  :color="status.color"
                  rounded
                  height="8"
                  class="overview-status-row__bar"
                />

                <div class="overview-status-row__footer">
                  <span>{{ status.percentageLabel }}</span>
                </div>
              </div>
            </div>
          </v-card>

          <v-card class="overview-panel mt-4" outlined>
            <div class="overview-panel__title">{{ $t('overview_top_templates') }}</div>
            <div class="overview-panel__subtitle">{{ $t('overview_top_templates_subtitle') }}</div>

            <div v-if="topTemplates.length === 0" class="overview-empty-state">
              {{ $t('overview_no_templates') }}
            </div>

            <div
              v-for="item in topTemplates"
              :key="item.templateKey"
              class="overview-template-row"
            >
              <div class="overview-template-row__icon">
                <v-icon small>{{ getAppIcon(item.tpl_app) }}</v-icon>
              </div>

              <div class="overview-template-row__body">
                <router-link
                  class="overview-template-row__name"
                  :to="`/project/${projectId}/templates/${item.template_id}`"
                >
                  {{ item.tpl_alias }}
                </router-link>
                <div class="overview-template-row__meta">
                  {{ $t('overview_runs_count', { count: item.count }) }}
                </div>
              </div>

              <div class="overview-template-row__time">
                {{ item.lastStarted | formatDate }}
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>

      <v-card class="overview-panel mt-4" outlined>
        <div class="overview-panel__header">
          <div>
            <div class="overview-panel__title">{{ $t('overview_recent_runs') }}</div>
            <div class="overview-panel__subtitle">{{ $t('overview_recent_runs_subtitle') }}</div>
          </div>

          <div class="overview-updated-at" v-if="lastUpdatedAt">
            {{ $t('overview_last_updated', { time: lastUpdatedLabel }) }}
          </div>
        </div>

        <v-data-table
          :headers="headers"
          :items="recentTableItems"
          hide-default-footer
          class="overview-table"
          :items-per-page="8"
        >
          <template v-slot:item.tpl_alias="{ item }">
            <div class="overview-task">
              <div class="overview-task__title">
                <v-icon
                  class="mr-2"
                  small
                >
                  {{ getAppIcon(item.tpl_app) }}
                </v-icon>

                <TaskLink
                  :task-id="item.id"
                  :label="'#' + item.id"
                />

                <v-icon small class="mx-1">mdi-arrow-right</v-icon>

                <router-link :to="`/project/${item.project_id}/templates/${item.template_id}`">
                  {{ item.tpl_alias }}
                </router-link>
              </div>

              <div v-if="item.message || item.commit_hash" class="overview-task__meta">
                <span v-if="item.message">
                  <v-icon x-small>mdi-message-outline</v-icon>
                  {{ item.message }}
                </span>
                <span v-else>
                  <v-icon x-small>mdi-source-fork</v-icon>
                  {{ item.commit_message }}
                </span>
              </div>
            </div>
          </template>

          <template v-slot:item.status="{ item }">
            <TaskStatus :status="item.status" />
          </template>

          <template v-slot:item.start="{ item }">
            {{ (item.start || item.created) | formatDate }}
          </template>

          <template v-slot:item.end="{ item }">
            {{ [item.start || item.created, item.end] | formatMilliseconds }}
          </template>
        </v-data-table>
      </v-card>
    </div>
  </div>
</template>

<style lang="scss">
.overview-shell {
  max-width: calc(var(--breakpoint-xl) - var(--nav-drawer-width) - 56px);
  margin: auto;
}

.overview-loading {
  border-radius: 999px;
}

.overview-toolbar {
  display: flex;
  justify-content: flex-end;
}

.overview-toolbar__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: flex-end;
}

.overview-range-toggle {
  background: transparent !important;
  box-shadow: none;
  padding: 0 !important;
}

.overview-range-toggle .v-btn {
  background: transparent !important;
  border: 1px solid rgba(96, 125, 139, 0.26);
  border-radius: 12px !important;
  box-shadow: none !important;
  color: #546e7a !important;
  min-width: 46px !important;
  text-transform: none;
}

.overview-range-toggle .v-btn.v-btn--active {
  background: rgba(56, 131, 255, 0.12) !important;
  border-color: rgba(56, 131, 255, 0.38);
  color: #1976d2 !important;
}

.overview-user-filter {
  max-width: 220px;
  min-width: 220px;
}

.overview-user-filter .v-input__slot {
  background: rgba(255, 255, 255, 0.72) !important;
  border-radius: 14px !important;
}

.overview-metric-card,
.overview-panel {
  border-radius: 22px !important;
}

.overview-metric-card {
  height: 100%;
  padding: 20px;
}

.overview-metric-card__top {
  align-items: center;
  display: flex;
  justify-content: space-between;
}

.overview-metric-card__icon {
  align-items: center;
  border-radius: 14px;
  display: inline-flex;
  height: 42px;
  justify-content: center;
  width: 42px;
}

.overview-metric-card__icon--positive {
  background: rgba(46, 125, 50, 0.12);
  color: #2e7d32;
}

.overview-metric-card__icon--negative {
  background: rgba(211, 47, 47, 0.12);
  color: #d32f2f;
}

.overview-metric-card__icon--neutral {
  background: rgba(0, 80, 87, 0.10);
  color: #005057;
}

.overview-metric-card__label {
  color: #607d8b;
  font-size: 13px;
  margin-top: 18px;
}

.overview-metric-card__value {
  color: #102027;
  font-size: 32px;
  font-weight: 700;
  letter-spacing: -0.04em;
  margin-top: 8px;
}

.overview-metric-card__caption {
  color: #78909c;
  font-size: 12px;
  margin-top: 10px;
}

.overview-delta-chip {
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  padding: 6px 10px;
}

.overview-delta-chip--positive {
  background: rgba(46, 125, 50, 0.12);
  color: #2e7d32;
}

.overview-delta-chip--negative {
  background: rgba(211, 47, 47, 0.12);
  color: #d32f2f;
}

.overview-delta-chip--neutral {
  background: rgba(96, 125, 139, 0.12);
  color: #607d8b;
}

.overview-panel {
  padding: 22px;
}

.overview-panel__header {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.overview-panel__title {
  color: #102027;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.overview-panel__subtitle {
  color: #607d8b;
  font-size: 13px;
  margin-top: 4px;
}

.overview-panel__summary {
  text-align: right;
}

.overview-panel__summary-value {
  color: #102027;
  font-size: 28px;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.overview-panel__summary-label,
.overview-updated-at {
  color: #78909c;
  font-size: 12px;
}

.overview-chart-wrapper {
  height: 340px;
  margin-top: 20px;
}

.overview-status-list {
  margin-top: 18px;
}

.overview-status-row + .overview-status-row {
  margin-top: 18px;
}

.overview-status-row__header,
.overview-status-row__footer {
  color: #455a64;
  display: flex;
  font-size: 13px;
  justify-content: space-between;
}

.overview-status-row__bar {
  margin-top: 8px;
}

.overview-status-row__footer {
  color: #78909c;
  margin-top: 6px;
}

.overview-template-row {
  align-items: center;
  display: flex;
  gap: 14px;
  padding: 14px 0;
}

.overview-template-row + .overview-template-row {
  border-top: 1px solid rgba(96, 125, 139, 0.14);
}

.overview-template-row__icon {
  align-items: center;
  background: rgba(0, 80, 87, 0.08);
  border-radius: 12px;
  color: #005057;
  display: inline-flex;
  height: 38px;
  justify-content: center;
  width: 38px;
}

.overview-template-row__body {
  flex: 1;
  min-width: 0;
}

.overview-template-row__name {
  display: inline-block;
  font-weight: 600;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.overview-template-row__meta,
.overview-template-row__time,
.overview-empty-state,
.overview-task__meta {
  color: #78909c;
  font-size: 12px;
}

.overview-empty-state {
  margin-top: 18px;
}

.overview-table {
  margin-top: 16px;
}

.overview-table .v-data-table__wrapper table tbody tr td {
  padding-bottom: 16px;
  padding-top: 16px;
}

.overview-task__title {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}

.overview-task__meta {
  margin-left: 28px;
  margin-top: 6px;
}

.theme--dark {
  .overview-template-row__icon {
    color: #80cbc4;
  }

  .overview-metric-card__value,
  .overview-panel__title,
  .overview-panel__summary-value {
    color: #eceff1;
  }

  .overview-status-row__header {
    color: #b0bec5;
  }

  .overview-metric-card__label,
  .overview-panel__subtitle,
  .overview-panel__summary-label,
  .overview-updated-at,
  .overview-status-row__footer,
  .overview-template-row__meta,
  .overview-template-row__time,
  .overview-empty-state,
  .overview-task__meta,
  .overview-metric-card__caption {
    color: #90a4ae;
  }

  .overview-template-row + .overview-template-row {
    border-top-color: rgba(255, 255, 255, 0.08);
  }

  .overview-range-toggle .v-btn {
    border-color: rgba(255, 255, 255, 0.12);
    color: #b0bec5 !important;
  }

  .overview-range-toggle .v-btn.v-btn--active {
    background: rgba(77, 162, 255, 0.16) !important;
    border-color: rgba(77, 162, 255, 0.42);
    color: #90caf9 !important;
  }

  .overview-user-filter .v-input__slot {
    background: rgba(12, 21, 34, 0.52) !important;
  }
}

@media (max-width: 1264px) {
  .overview-shell {
    max-width: calc(100vw - 32px);
  }
}

@media (max-width: 960px) {
  .overview-shell {
    max-width: 100%;
  }

  .overview-toolbar,
  .overview-panel__header {
    flex-direction: column;
  }

  .overview-toolbar__actions {
    justify-content: flex-start;
    width: 100%;
  }

  .overview-user-filter {
    max-width: none;
    min-width: 0;
    width: 100%;
  }

  .overview-chart-wrapper {
    height: 300px;
  }
}
</style>

<script>
import axios from 'axios';
import dayjs from 'dayjs';
import EventBus from '@/event-bus';
import DashboardMenu from '@/components/DashboardMenu.vue';
import OverviewTrendChart from '@/components/OverviewTrendChart.vue';
import TaskLink from '@/components/TaskLink.vue';
import TaskStatus from '@/components/TaskStatus.vue';
import socket from '@/socket';
import AppsMixin from '@/components/AppsMixin';
import PermissionsCheck from '@/components/PermissionsCheck';
import { USER_PERMISSIONS } from '@/lib/constants';

const ACTIVE_STATUSES = ['waiting', 'starting', 'running', 'stopping', 'waiting_confirmation', 'confirmed'];

export default {
  components: {
    DashboardMenu,
    OverviewTrendChart,
    TaskLink,
    TaskStatus,
  },

  mixins: [AppsMixin, PermissionsCheck],

  props: {
    projectId: Number,
    projectType: String,
    userId: Number,
    userRole: String,
    user: Object,
  },

  data() {
    return {
      headers: [{
        text: this.$i18n.t('task2'),
        value: 'tpl_alias',
        sortable: false,
      }, {
        text: this.$i18n.t('status'),
        value: 'status',
        sortable: false,
      }, {
        text: this.$i18n.t('user'),
        value: 'user_name',
        sortable: false,
      }, {
        text: this.$i18n.t('start'),
        value: 'start',
        sortable: false,
      }, {
        text: this.$i18n.t('duration'),
        value: 'end',
        sortable: false,
      }],
      dateRange: '30d',
      selectedUser: null,
      rawUsers: [],
      allTasks: [],
      isLoading: false,
      lastUpdatedAt: null,
      socketListenerId: null,
      tasksRefreshTimer: null,
    };
  },

  computed: {
    USER_PERMISSIONS() {
      return USER_PERMISSIONS;
    },

    dateRanges() {
      return [{
        label: '7D',
        value: '7d',
        days: 7,
      }, {
        label: '30D',
        value: '30d',
        days: 30,
      }, {
        label: '90D',
        value: '90d',
        days: 90,
      }, {
        label: '365D',
        value: '365d',
        days: 365,
      }];
    },

    users() {
      return [{
        text: this.$t('all_users'),
        value: null,
      }, ...this.rawUsers];
    },

    rangeDays() {
      const selectedRange = this.dateRanges.find((range) => range.value === this.dateRange);
      return (selectedRange || this.dateRanges[1]).days;
    },

    currentStartDate() {
      return dayjs().subtract(this.rangeDays - 1, 'day').format('YYYY-MM-DD');
    },

    currentEndDate() {
      return dayjs().add(1, 'day').format('YYYY-MM-DD');
    },

    previousStartDate() {
      return dayjs(this.currentStartDate).subtract(this.rangeDays, 'day').format('YYYY-MM-DD');
    },

    previousEndDate() {
      return this.currentStartDate;
    },

    userFilteredTasks() {
      return this.allTasks.filter((task) => (
        this.selectedUser === null
          || this.selectedUser === undefined
          || task.user_id === this.selectedUser
      ));
    },

    sortedTasks() {
      return [...this.userFilteredTasks].sort(
        (a, b) => this.getTaskDateValue(b) - this.getTaskDateValue(a),
      );
    },

    currentRangeTasks() {
      return this.filterTasksByRange(
        this.sortedTasks,
        this.currentStartDate,
        this.currentEndDate,
      );
    },

    previousRangeTasks() {
      return this.filterTasksByRange(
        this.sortedTasks,
        this.previousStartDate,
        this.previousEndDate,
      );
    },

    normalizedStats() {
      return this.normalizeTasksByDate(
        this.currentRangeTasks,
        this.currentStartDate,
        this.rangeDays,
      );
    },

    currentTotals() {
      return this.getTotals(this.currentRangeTasks);
    },

    previousTotals() {
      return this.getTotals(this.previousRangeTasks);
    },

    totalRuns() {
      return this.currentTotals.total;
    },

    successRate() {
      return this.getRate(this.currentTotals.success, this.currentTotals.finished);
    },

    recentTableItems() {
      return this.sortedTasks.slice(0, 8);
    },

    topTemplates() {
      const aggregated = this.currentRangeTasks
        .reduce((acc, task) => {
          const templateKey = `${task.template_id}:${task.tpl_alias}`;
          const current = acc[templateKey] || {
            templateKey,
            template_id: task.template_id,
            tpl_alias: task.tpl_alias,
            tpl_app: task.tpl_app,
            count: 0,
            lastStarted: task.start || task.created,
          };

          current.count += 1;

          if (dayjs(task.start || task.created).isAfter(dayjs(current.lastStarted))) {
            current.lastStarted = task.start || task.created;
          }

          acc[templateKey] = current;
          return acc;
        }, {});

      return Object.values(aggregated)
        .sort((a, b) => {
          if (b.count !== a.count) {
            return b.count - a.count;
          }

          return dayjs(b.lastStarted).valueOf() - dayjs(a.lastStarted).valueOf();
        })
        .slice(0, 5);
    },

    statusBreakdown() {
      return [{
        key: 'success',
        label: this.$t('status_success'),
        color: '#2e7d32',
        count: this.currentTotals.success,
      }, {
        key: 'error',
        label: this.$t('status_failed'),
        color: '#d32f2f',
        count: this.currentTotals.error,
      }, {
        key: 'active',
        label: this.$t('overview_active_runs'),
        color: '#1e88e5',
        count: this.currentTotals.active,
      }, {
        key: 'stopped',
        label: this.$t('status_stopped'),
        color: '#607d8b',
        count: this.currentTotals.stopped,
      }].map((status) => ({
        ...status,
        percentage: this.getRate(status.count, this.totalRuns),
        percentageLabel: `${this.getRate(status.count, this.totalRuns).toFixed(1)}%`,
        countLabel: this.formatNumber(status.count),
      }));
    },

    activeRunsCount() {
      return this.sortedTasks.filter((task) => ACTIVE_STATUSES.includes(task.status)).length;
    },

    summaryCards() {
      return [{
        title: this.$t('overview_total_runs'),
        value: this.formattedTotalRuns,
        icon: 'mdi-lightning-bolt-outline',
        tone: 'neutral',
        delta: this.getPercentDelta(this.totalRuns, this.previousTotals.total),
      }, {
        title: this.$t('overview_success_rate'),
        value: `${this.successRate.toFixed(1)}%`,
        icon: 'mdi-check-decagram-outline',
        tone: 'positive',
        delta: this.getPointsDelta(
          this.successRate,
          this.getRate(this.previousTotals.success, this.previousTotals.finished),
        ),
      }, {
        title: this.$t('overview_failed_runs'),
        value: this.formatNumber(this.currentTotals.error),
        icon: 'mdi-alert-circle-outline',
        tone: 'negative',
        delta: this.getPercentDelta(this.currentTotals.error, this.previousTotals.error, true),
      }, {
        title: this.$t('overview_active_runs'),
        value: this.formatNumber(this.activeRunsCount),
        icon: 'mdi-progress-clock',
        tone: 'neutral',
        delta: {
          text: this.$t('overview_live'),
          tone: 'neutral',
        },
      }];
    },

    formattedTotalRuns() {
      return this.formatNumber(this.totalRuns);
    },

    lastRunTask() {
      return this.sortedTasks[0] || null;
    },

    lastRunStartedAt() {
      if (!this.lastRunTask) {
        return null;
      }

      return this.lastRunTask.start || this.lastRunTask.created;
    },

    lastUpdatedLabel() {
      if (!this.lastUpdatedAt) {
        return '';
      }

      return dayjs(this.lastUpdatedAt).fromNow();
    },
  },

  watch: {
    async projectId() {
      this.selectedUser = null;
      await this.refreshAll();
    },
  },

  async created() {
    await this.refreshAll();
    this.socketListenerId = socket.addListener((data) => this.onWebsocketDataReceived(data));
  },

  beforeDestroy() {
    if (this.socketListenerId) {
      socket.removeListener(this.socketListenerId);
    }

    if (this.tasksRefreshTimer) {
      clearTimeout(this.tasksRefreshTimer);
    }
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    async refreshAll() {
      this.isLoading = true;

      try {
        await Promise.all([
          this.refreshTasks(),
          this.loadUsers(),
        ]);
      } finally {
        this.isLoading = false;
      }
    },

    async refreshTasks() {
      this.allTasks = (await axios({
        method: 'get',
        url: `/api/project/${this.projectId}/tasks`,
        responseType: 'json',
      })).data;

      this.lastUpdatedAt = new Date();
    },

    async loadUsers() {
      try {
        this.rawUsers = (await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/users`,
          responseType: 'json',
        })).data.map((user) => ({
          value: user.id,
          text: user.name,
        }));
      } catch (e) {
        this.rawUsers = [];
      }
    },

    onWebsocketDataReceived(data) {
      if (data.project_id !== this.projectId || data.type !== 'update') {
        return;
      }

      const existingTask = this.allTasks.find((task) => task.id === data.task_id);

      if (existingTask) {
        Object.assign(existingTask, {
          ...data,
          type: undefined,
        });
      } else {
        this.scheduleTasksRefresh();
      }

      this.lastUpdatedAt = new Date();
      this.scheduleTasksRefresh();
    },

    scheduleTasksRefresh() {
      if (this.tasksRefreshTimer) {
        clearTimeout(this.tasksRefreshTimer);
      }

      this.tasksRefreshTimer = setTimeout(() => {
        this.refreshTasks();
      }, 800);
    },

    filterTasksByRange(tasks, startDate, endDate) {
      const startValue = dayjs(startDate).startOf('day').valueOf();
      const endValue = dayjs(endDate).startOf('day').valueOf();

      return tasks.filter((task) => {
        const taskDate = this.getTaskDateValue(task);
        return taskDate >= startValue && taskDate < endValue;
      });
    },

    normalizeTasksByDate(tasks, startDate, days) {
      const statsByDate = (tasks || []).reduce((acc, task) => {
        const date = dayjs(this.getTaskDate(task)).format('YYYY-MM-DD');

        if (!acc[date]) {
          acc[date] = {
            success: 0,
            error: 0,
            stopped: 0,
            active: 0,
          };
        }

        if (task.status === 'success') {
          acc[date].success += 1;
        } else if (task.status === 'error') {
          acc[date].error += 1;
        } else if (task.status === 'stopped') {
          acc[date].stopped += 1;
        } else if (ACTIVE_STATUSES.includes(task.status)) {
          acc[date].active += 1;
        }

        return acc;
      }, {});

      return Array.from({ length: days }, (_, index) => {
        const date = dayjs(startDate).add(index, 'day').format('YYYY-MM-DD');
        const countByStatus = statsByDate[date] || {
          success: 0,
          error: 0,
          stopped: 0,
          active: 0,
        };

        return {
          date,
          count_by_status: countByStatus,
        };
      });
    },

    getTotals(rows) {
      const totals = rows.reduce((acc, task) => {
        acc.total += 1;

        if (task.status === 'success') {
          acc.success += 1;
          acc.finished += 1;
        } else if (task.status === 'error') {
          acc.error += 1;
          acc.finished += 1;
        } else if (task.status === 'stopped') {
          acc.stopped += 1;
          acc.finished += 1;
        } else if (ACTIVE_STATUSES.includes(task.status)) {
          acc.active += 1;
        }

        return acc;
      }, {
        total: 0,
        success: 0,
        error: 0,
        stopped: 0,
        active: 0,
        finished: 0,
      });

      return totals;
    },

    getRate(part, total) {
      if (!total) {
        return 0;
      }

      return (part / total) * 100;
    },

    formatNumber(value, maximumFractionDigits = 0) {
      return new Intl.NumberFormat(undefined, {
        minimumFractionDigits: maximumFractionDigits > 0 ? 1 : 0,
        maximumFractionDigits,
      }).format(value || 0);
    },

    getPercentDelta(current, previous, reverse = false) {
      if (current === 0 && previous === 0) {
        return {
          text: this.$t('overview_no_change'),
          tone: 'neutral',
        };
      }

      if (previous === 0) {
        return {
          text: this.$t('overview_new_data'),
          tone: reverse ? 'negative' : 'positive',
        };
      }

      const delta = ((current - previous) / previous) * 100;

      return {
        text: `${delta > 0 ? '+' : ''}${delta.toFixed(Math.abs(delta) < 10 ? 1 : 0)}%`,
        tone: this.getDeltaTone(delta, reverse),
      };
    },

    getPointsDelta(current, previous) {
      const delta = current - previous;

      if (delta === 0) {
        return {
          text: this.$t('overview_no_change'),
          tone: 'neutral',
        };
      }

      return {
        text: `${delta > 0 ? '+' : ''}${delta.toFixed(1)} pp`,
        tone: this.getDeltaTone(delta),
      };
    },

    getDeltaTone(delta, reverse = false) {
      if (delta === 0) {
        return 'neutral';
      }

      const isPositive = reverse ? delta < 0 : delta > 0;
      return isPositive ? 'positive' : 'negative';
    },

    getTaskDate(task) {
      return task.start || task.end || task.created;
    },

    getTaskDateValue(task) {
      return dayjs(this.getTaskDate(task)).valueOf() || 0;
    },
  },
};
</script>
