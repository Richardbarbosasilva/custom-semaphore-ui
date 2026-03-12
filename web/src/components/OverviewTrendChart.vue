<template>
  <LineChartGenerator
    :chart-id="chartId"
    :dataset-id-key="chartId"
    :chart-options="chartOptions"
    :chart-data="chartData"
  />
</template>

<script>
import {
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  TimeScale,
  Title,
  Tooltip,
} from 'chart.js';
import './chartjs-adapter-day';

ChartJS.register(
  Title,
  Tooltip,
  Legend,
  Filler,
  LineElement,
  LinearScale,
  PointElement,
  TimeScale,
);

export default {
  name: 'OverviewTrendChart',

  props: {
    chartId: {
      type: String,
      default: 'overview-trend-chart',
    },
    sourceData: {
      type: Array,
      default: () => [],
    },
    dark: Boolean,
  },

  computed: {
    axisColor() {
      return this.dark ? '#cfd8dc' : '#546e7a';
    },

    gridColor() {
      return this.dark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(38, 50, 56, 0.08)';
    },

    chartData() {
      return {
        labels: this.sourceData.map((row) => new Date(row.date)),
        datasets: [{
          label: this.$t('status_success'),
          data: this.sourceData.map((row) => row.count_by_status.success),
          borderColor: '#2e7d32',
          backgroundColor: 'rgba(46, 125, 50, 0.18)',
          fill: true,
          pointRadius: 0,
          pointHoverRadius: 4,
          tension: 0.35,
          borderWidth: 2,
        }, {
          label: this.$t('status_failed'),
          data: this.sourceData.map((row) => row.count_by_status.error),
          borderColor: '#d32f2f',
          backgroundColor: 'rgba(211, 47, 47, 0.10)',
          fill: true,
          pointRadius: 0,
          pointHoverRadius: 4,
          tension: 0.35,
          borderWidth: 2,
        }, {
          label: this.$t('overview_active_runs'),
          data: this.sourceData.map((row) => row.count_by_status.active),
          borderColor: '#1e88e5',
          backgroundColor: 'rgba(30, 136, 229, 0.08)',
          fill: false,
          pointRadius: 0,
          pointHoverRadius: 4,
          tension: 0.35,
          borderWidth: 2,
          borderDash: [10, 6],
        }, {
          label: this.$t('status_stopped'),
          data: this.sourceData.map((row) => row.count_by_status.stopped),
          borderColor: '#607d8b',
          backgroundColor: 'rgba(96, 125, 139, 0.08)',
          fill: true,
          pointRadius: 0,
          pointHoverRadius: 4,
          tension: 0.35,
          borderWidth: 2,
          borderDash: [6, 6],
        }],
      };
    },

    chartOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false,
        },
        plugins: {
          legend: {
            position: 'top',
            align: 'start',
            labels: {
              color: this.axisColor,
              usePointStyle: true,
              boxWidth: 8,
            },
          },
          tooltip: {
            backgroundColor: this.dark ? '#102027' : '#ffffff',
            titleColor: this.dark ? '#ffffff' : '#102027',
            bodyColor: this.dark ? '#eceff1' : '#37474f',
            borderColor: this.gridColor,
            borderWidth: 1,
          },
        },
        scales: {
          x: {
            type: 'time',
            time: {
              unit: this.sourceData.length > 45 ? 'month' : 'day',
              tooltipFormat: 'll',
            },
            grid: {
              display: false,
            },
            ticks: {
              color: this.axisColor,
              maxTicksLimit: 8,
            },
          },
          y: {
            beginAtZero: true,
            grid: {
              color: this.gridColor,
            },
            ticks: {
              color: this.axisColor,
              precision: 0,
            },
          },
        },
      };
    },
  },
};
</script>
