<template>
  <v-card class="task-stats" outlined>
    <div class="task-stats__header">
      <div>
        <div class="task-stats__title">{{ $t('project_stats') }}</div>
        <div class="task-stats__subtitle">{{ $t('stats_visual_subtitle') }}</div>
      </div>

      <div class="task-stats__filters">
        <v-btn-toggle
          v-model="dateRange"
          mandatory
          tile
          class="task-stats__toggle"
        >
          <v-btn
            v-for="range in dateRanges"
            :key="range.value"
            :value="range.value"
            small
            depressed
          >
            {{ range.text }}
          </v-btn>
        </v-btn-toggle>

        <v-select
          hide-details
          dense
          outlined
          :items="users"
          class="task-stats__select"
          v-model="user"
        />
      </div>
    </div>

    <v-card-text class="task-stats__body">
      <LineChart :source-data="stats"/>
    </v-card-text>
  </v-card>
</template>
<style lang="scss">
.task-stats {
  border-radius: 24px !important;
  padding: 24px;
}

.task-stats__header {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.task-stats__title {
  color: #102027;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.task-stats__subtitle {
  color: #607d8b;
  font-size: 13px;
  margin-top: 4px;
  max-width: 460px;
}

.task-stats__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: flex-end;
}

.task-stats__toggle {
  background: transparent !important;
  box-shadow: none !important;
}

.task-stats__toggle .v-btn {
  background: transparent !important;
  border: 1px solid rgba(96, 125, 139, 0.26);
  border-radius: 12px !important;
  box-shadow: none !important;
  color: #546e7a !important;
  min-width: 46px !important;
  text-transform: none;
}

.task-stats__toggle .v-btn.v-btn--active {
  background: rgba(56, 131, 255, 0.12) !important;
  border-color: rgba(56, 131, 255, 0.38);
  color: #1976d2 !important;
}

.task-stats__select {
  min-width: 220px;
}

.task-stats__select .v-input__slot {
  background: rgba(255, 255, 255, 0.72) !important;
  border-radius: 14px !important;
}

.task-stats__body {
  margin-top: 16px;
  padding: 0 !important;
}

.task-stats__body > div {
  height: 360px;
}

.theme--dark {
  .task-stats__title {
    color: #eceff1;
  }

  .task-stats__subtitle {
    color: #90a4ae;
  }

  .task-stats__toggle .v-btn {
    border-color: rgba(255, 255, 255, 0.12);
    color: #b0bec5 !important;
  }

  .task-stats__toggle .v-btn.v-btn--active {
    background: rgba(77, 162, 255, 0.16) !important;
    border-color: rgba(77, 162, 255, 0.42);
    color: #90caf9 !important;
  }

  .task-stats__select .v-input__slot {
    background: rgba(12, 21, 34, 0.52) !important;
  }
}

@media (max-width: 960px) {
  .task-stats {
    padding: 18px;
  }

  .task-stats__header {
    flex-direction: column;
  }

  .task-stats__filters,
  .task-stats__select {
    width: 100%;
  }
}
</style>
<script>
import axios from 'axios';
import LineChart from '@/components/LineChart.vue';

export default {
  components: { LineChart },

  props: {
    templateId: Number,
    projectId: Number,
  },

  data() {
    return {
      dateRanges: [{
        text: '7D',
        value: 'last_week',
      }, {
        text: '30D',
        value: 'last_month',
      }, {
        text: '365D',
        value: 'last_year',
      }],
      users: [{
        text: this.$t('all_users'),
        value: null,
      }],
      user: null,
      stats: null,
      dateRange: 'last_week',
    };
  },

  computed: {
    startDate() {
      const date = new Date();

      switch (this.dateRange) {
        case 'last_year':
          date.setFullYear(date.getFullYear() - 1);
          break;
        case 'last_month':
          date.setDate(date.getDate() - 30);
          break;
        case 'last_week':
        default:
          date.setDate(date.getDate() - 7);
          break;
      }

      return date.toISOString().split('T')[0];
    },
  },

  watch: {
    async startDate() {
      await this.refreshData();
    },
    async user() {
      await this.refreshData();
    },
  },

  async created() {
    await this.refreshData();

    this.users = [{
      text: this.$t('all_users'),
      value: null,
    }, ...(await axios({
      method: 'get',
      url: `/api/project/${this.projectId}/users`,
      responseType: 'json',
    })).data.map((x) => ({
      value: x.id,
      text: x.name,
    }))];
  },

  methods: {
    async refreshData() {
      let url;

      if (this.templateId) {
        url = `/api/project/${this.projectId}/templates/${this.templateId}/stats?start=${this.startDate}`;
      } else {
        url = `/api/project/${this.projectId}/stats?start=${this.startDate}`;
      }

      if (this.user) {
        url += `&user_id=${this.user}`;
      }

      this.stats = (await axios({
        method: 'get',
        url,
        responseType: 'json',
      })).data;

      const firstPoint = this.stats[0];

      if (!firstPoint || firstPoint.date > this.startDate) {
        this.stats.unshift({
          date: this.startDate,
          count_by_status: {
            success: 0,
            failed: 0,
            stopped: 0,
          },
        });
      }

      const lastPoint = this.stats[this.stats.length - 1];

      if (lastPoint.date < new Date().toISOString().split('T')[0]) {
        this.stats.push({
          date: new Date().toISOString().split('T')[0],
          count_by_status: {
            success: 0,
            failed: 0,
            stopped: 0,
          },
        });
      }
    },
  },
};
</script>
