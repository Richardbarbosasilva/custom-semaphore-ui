<template>
  <div>

    <EditDialog
      v-model="passwordDialog"
      save-button-text="Save"
      :title="$t('changePassword')"
      v-if="item"
      event-name="i-user"
    >
      <template v-slot:form="{ onSave, onError, needSave, needReset }">
        <ChangePasswordForm
          :project-id="projectId"
          :item-id="item.id"
          @save="onSave"
          @error="onError"
          :need-save="needSave"
          :need-reset="needReset"
        />
      </template>
    </EditDialog>

    <v-tabs v-model="tab">
      <v-tab key="settings">Settings</v-tab>
      <v-tab
        key="2fa"
        v-if="showSecurityTab"
      >
        Security
      </v-tab>
    </v-tabs>

    <v-divider class="mb-6" style="margin-top: -1px;"/>

    <v-tabs-items v-model="tab" style="overflow: unset;">
      <v-tab-item key="settings">
        <v-form
          ref="form"
          lazy-validation
          v-model="formValid"
          v-if="item != null"
        >
          <v-alert
            :value="formError"
            color="error"
            class="pb-2"
          >{{ formError }}
          </v-alert>

          <v-text-field
            v-model="item.name"
            :label="$t('name')"
            :rules="[v => !!v || $t('name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.username"
            :label="$t('username')"
            :rules="[v => !!v || $t('user_name_required')]"
            required
            :disabled="formSaving"
            outlined
            dense
          ></v-text-field>

          <v-text-field
            v-model="item.email"
            :label="$t('email')"
            :rules="[v => !!v || $t('email_required')]"
            required
            :disabled="!isNew && item.external || formSaving"
            outlined
            dense
          >

            <template v-slot:append>
              <v-chip outlined color="green" disabled small style="opacity: 1">private</v-chip>
            </template>
          </v-text-field>

          <v-text-field
            v-if="isNew"
            v-model="item.password"
            :label="$t('password')"
            class="masked-secret-input"
            :required="isNew && !item.external"
            :rules="isNew && !item.external ? [v => !!v || $t('password_required')] : []"
            :disabled="item.external || formSaving"
            outlined
            dense
          ></v-text-field>

          <v-row class="pb-5 pt-2">
            <v-col cols="6">
              <v-checkbox
                dense
                hide-details
                v-model="item.alert"
                :label="$t('sendAlerts')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                dense
                hide-details
                v-model="item.admin"
                :label="$t('adminUser')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isPro">
              <v-checkbox
                :disabled="!isAdmin"
                dense
                hide-details
                v-model="item.pro"
                :label="$t('Pro user')"
              ></v-checkbox>
            </v-col>
            <v-col cols="6" v-if="isAdmin">
              <v-checkbox
                :disabled="!isNew"
                dense
                hide-details
                v-model="item.external"
                :label="$t('external')"
              ></v-checkbox>
            </v-col>
          </v-row>
        </v-form>
      </v-tab-item>

      <v-tab-item key="2fa" v-if="item != null && showSecurityTab">

        <div v-if="!isNew">
          <div class="title mb-3">Password</div>
          <v-btn color="primary" @click="passwordDialog = true;">Change password</v-btn>
        </div>

        <div
          :class="{'pt-10': !isNew}"
          v-if="authMethods.totp"
        >
          <div class="title mb-2">Two-factor authentication</div>

          <v-alert
            dense
            text
            color="info"
            class="mb-4"
          >
            O QR code nao e enviado por email. Ative o TOTP neste usuario,
            escaneie o QR code ou copie a chave manual e guarde o recovery code
            antes de sair da sessao.
          </v-alert>

          <v-switch
            class="mt-0"
            v-model="totpEnabled"
            label="Time-based one-time password"
          ></v-switch>

          <v-btn
            v-if="item.totp != null"
            small
            text
            color="primary"
            class="mb-4 px-0"
            @click="totpSetupDialog = true"
          >
            Abrir instrucoes de configuracao
          </v-btn>

          <img
            v-if="totpQrUrl"
            :src="totpQrUrl"
            style="
        width: 100%;
        aspect-ratio: 1;
        border-radius: 4px;
        display: block;
        margin: 0 auto 10px auto;
        border: 10px solid white;
        background-color: white;
      "
            alt="QR code"
          />

          <div
            v-if="totpSecret"
            class="mt-5 pb-3"
          >
            <div class="subtitle-1 mb-2">Manual setup key</div>
            <div style="position: relative;">
              <code style="font-size: 16px; background-color: #183149; color: #ffffff;">
                {{ totpSecret }}
              </code>

              <CopyClipboardButton
                style="position: absolute; right: -4px; top: -12px;"
                :text="totpSecret"
                large
                color="white"
              />
            </div>
          </div>

          <div
            v-if="authMethods.totp.allow_recovery && item.totp && item.totp.recovery_code"
            class="mt-5 pb-3"
          >
            <div class="subtitle-1 mb-2">Recovery code</div>
            <div style="position: relative;">
              <code
                style="font-size: 18px; background-color: #e03755;"
              >
                {{ item.totp.recovery_code }}
              </code>

              <CopyClipboardButton
                style="position: absolute; right: -4px; top: -12px;"
                :text="item.totp.recovery_code"
                large
                color="white"
              />
            </div>
          </div>

          <v-alert
            v-else-if="item.totp != null && authMethods.totp.allow_recovery"
            dense
            text
            color="warning"
            class="mt-4 mb-0"
          >
            O recovery code so e exibido no momento da ativacao. Se ele nao foi
            salvo, um administrador precisara desativar e habilitar o TOTP novamente.
          </v-alert>
        </div>
      </v-tab-item>
    </v-tabs-items>

    <v-dialog
      v-model="totpSetupDialog"
      max-width="540"
      v-if="item && item.totp"
    >
      <v-card class="pa-3">
        <v-card-title class="pb-2">
          Configuracao do TOTP
        </v-card-title>

        <v-card-text>
          <v-alert dense text color="info" class="mb-4">
            Use Google Authenticator, Authy ou app equivalente. Nada e enviado por email.
          </v-alert>

          <img
            v-if="totpQrUrl"
            :src="totpQrUrl"
            style="
              width: 100%;
              max-width: 280px;
              aspect-ratio: 1;
              border-radius: 10px;
              display: block;
              margin: 0 auto 16px auto;
              border: 10px solid white;
              background-color: white;
            "
            alt="QR code"
          />

          <div v-if="totpSecret" class="mb-5">
            <div class="subtitle-1 mb-2">Chave manual</div>
            <div style="position: relative;">
              <code style="font-size: 16px; background-color: #183149; color: #ffffff;">
                {{ totpSecret }}
              </code>

              <CopyClipboardButton
                style="position: absolute; right: -4px; top: -12px;"
                :text="totpSecret"
                large
                color="white"
              />
            </div>
          </div>

          <div
            v-if="authMethods.totp.allow_recovery && item.totp.recovery_code"
            class="mb-3"
          >
            <div class="subtitle-1 mb-2">Recovery code</div>
            <div style="position: relative;">
              <code style="font-size: 18px; background-color: #e03755;">
                {{ item.totp.recovery_code }}
              </code>

              <CopyClipboardButton
                style="position: absolute; right: -4px; top: -12px;"
                :text="item.totp.recovery_code"
                large
                color="white"
              />
            </div>
          </div>

          <v-alert dense text color="warning" class="mb-0">
            Termine o pareamento no aplicativo autenticador antes de encerrar a sessao.
          </v-alert>
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn text color="primary" @click="totpSetupDialog = false">
            Fechar
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
import ItemFormBase from '@/components/ItemFormBase';
import axios from 'axios';
import EditDialog from '@/components/EditDialog.vue';
import ChangePasswordForm from '@/components/ChangePasswordForm.vue';
import CopyClipboardButton from '@/components/CopyClipboardButton.vue';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

export default {
  components: { CopyClipboardButton, ChangePasswordForm, EditDialog },
  props: {
    isAdmin: Boolean,
    authMethods: Object,
    canManageSecurity: Boolean,
  },

  mixins: [ItemFormBase],

  data() {
    return {
      passwordDialog: null,
      totpEnabled: false,
      totpQrUrl: null,
      totpSetupDialog: false,

      tab: null,
    };
  },

  watch: {
    tab(value) {
      if (value === 0) {
        this.$emit('show-action-buttons');
      } else {
        this.$emit('hide-action-buttons');
      }
    },

    async totpEnabled(val) {
      try {
        if (val) {
          if (this.item.totp == null) {
            this.item.totp = (await axios({
              method: 'post',
              url: `/api/users/${this.itemId}/2fas/totp`,
              responseType: 'json',
            })).data;

            this.totpQrUrl = `${document.baseURI}api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
            this.totpSetupDialog = true;
          }
        } else if (this.item.totp != null) {
          await axios({
            method: 'delete',
            url: `/api/users/${this.itemId}/2fas/totp/${this.item.totp.id}`,
            responseType: 'json',
          });
          this.item.totp = null;
          this.totpQrUrl = null;
          this.totpSetupDialog = false;
        }
      } catch (error) {
        this.totpEnabled = this.item.totp != null;
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(error),
        });
      }
    },
  },

  computed: {

    isPro() {
      return (process.env.VUE_APP_BUILD_TYPE || '').startsWith('pro_');
    },

    showSecurityTab() {
      return this.canManageSecurity && (!this.isNew || this.authMethods.totp);
    },

    totpSecret() {
      if (!this.item || !this.item.totp || !this.item.totp.url) {
        return null;
      }

      const match = this.item.totp.url.match(/[?&]secret=([^&]+)/);
      return match ? decodeURIComponent(match[1]) : null;
    },

  },

  methods: {

    afterLoadData() {
      if (this.item.totp == null) {
        this.totpEnabled = false;
        this.totpQrUrl = null;
        this.totpSetupDialog = false;
      } else {
        this.totpEnabled = true;
        this.totpQrUrl = `${document.baseURI}api/users/${this.itemId}/2fas/totp/${this.item.totp.id}/qr`;
      }
    },

    getItemsUrl() {
      return '/api/users';
    },

    getSingleItemUrl() {
      return `/api/users/${this.itemId}`;
    },
  },
};
</script>
