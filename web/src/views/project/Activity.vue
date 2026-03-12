<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div v-if="items">
    <v-toolbar flat>
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="can(USER_PERMISSIONS.updateProject)"
    />

    <div class="activity-shell px-4 pb-6">
      <v-card class="activity-panel mt-4" outlined>
        <div class="activity-panel__header">
          <div>
            <div class="activity-panel__title">{{ $t('activity') }}</div>
            <div class="activity-panel__subtitle">{{ $t('activity_feed_subtitle') }}</div>
          </div>

          <v-chip small outlined class="activity-panel__chip">
            {{ items.length }}
          </v-chip>
        </div>

        <v-data-table
          :headers="headers"
          :items="items"
          class="mt-4"
          :footer-props="{ itemsPerPageOptions: [20] }"
        >
          <template v-slot:item.created="{ item }">
            {{ item.created | formatDate }}
          </template>
        </v-data-table>
      </v-card>
    </div>
  </div>
</template>
<style lang="scss">
.activity-shell {
  margin: auto;
  max-width: calc(var(--breakpoint-lg) - var(--nav-drawer-width));
}

.activity-panel {
  border-radius: 22px !important;
  padding: 22px;
}

.activity-panel__header {
  align-items: center;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.activity-panel__title {
  color: #102027;
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.activity-panel__subtitle {
  color: #607d8b;
  font-size: 13px;
  margin-top: 4px;
  max-width: 520px;
}

.activity-panel__chip {
  font-weight: 700;
}

.theme--dark {
  .activity-panel__title {
    color: #eceff1;
  }

  .activity-panel__subtitle {
    color: #90a4ae;
  }
}

@media (max-width: 1264px) {
  .activity-shell {
    max-width: calc(100vw - 32px);
  }
}

@media (max-width: 960px) {
  .activity-shell {
    max-width: 100%;
  }

  .activity-panel__header {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
<script>
import ItemListPageBase from '@/components/ItemListPageBase';
import DashboardMenu from '@/components/DashboardMenu.vue';

export default {
  components: { DashboardMenu },

  mixins: [ItemListPageBase],

  methods: {
    getHeaders() {
      return [
        {
          text: this.$i18n.t('time'),
          value: 'created',
          sortable: false,
          width: '20%',
        },
        {
          text: this.$i18n.t('user'),
          value: 'username',
          sortable: false,
          width: '10%',
        },
        {
          text: this.$i18n.t('description'),
          value: 'description',
          sortable: false,
          width: '70%',
        },
      ];
    },

    getItemsUrl() {
      return `/api/project/${this.projectId}/events/last`;
    },
  },
};
</script>
