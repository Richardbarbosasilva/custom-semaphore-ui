<template>
  <div class="auth-admin">
    <div v-if="loading" class="auth-admin__loading">
      <v-progress-circular indeterminate color="primary" size="44" />
    </div>

    <div v-else-if="loadError" class="auth-admin__loading">
      <v-card flat class="auth-admin__error-card">
        <div class="auth-admin__error-title">Falha ao carregar autenticação</div>
        <div class="auth-admin__error-copy">{{ loadError }}</div>
        <v-btn color="primary" class="mt-6" @click="loadSettings()">Tentar novamente</v-btn>
      </v-card>
    </div>

    <div v-else>
      <v-toolbar flat class="auth-admin__toolbar">
        <v-btn
          icon
          class="mr-4"
          @click="returnToProjects()"
        >
          <v-icon>mdi-arrow-left</v-icon>
        </v-btn>

        <div>
          <div class="auth-admin__eyebrow">SECURITY CONTROL</div>
          <div class="auth-admin__title">Autenticação</div>
        </div>

        <v-spacer></v-spacer>

        <v-chip
          small
          outlined
          label
          class="mr-3"
          color="primary"
        >
          {{ form.ldap.enabled ? 'LDAP / Active Directory habilitado' : 'Login local ativo' }}
        </v-chip>

        <v-btn
          color="primary"
          class="auth-admin__save-btn"
          :loading="saving"
          @click="saveSettings()"
        >
          Salvar alterações
        </v-btn>
      </v-toolbar>

      <section class="auth-admin__hero">
        <div>
          <div class="auth-admin__hero-kicker">Identity & Access</div>
          <h1 class="auth-admin__hero-title">Controle unificado de login e segundo fator</h1>
          <p class="auth-admin__hero-copy">
            Ajuste o diretório LDAP, valide o mapeamento de atributos e defina a
            política global de TOTP para todo o ambiente sem depender do compose.
          </p>
        </div>

        <div class="auth-admin__hero-badges">
          <div class="auth-admin__hero-badge">
            <span class="auth-admin__hero-badge-label">Fluxo</span>
            <strong>{{ form.ldap.enabled ? 'Local + LDAP' : 'Local' }}</strong>
          </div>

          <div class="auth-admin__hero-badge">
            <span class="auth-admin__hero-badge-label">TOTP</span>
            <strong>{{ form.totp.enabled ? 'Obrigação disponível' : 'Desabilitado' }}</strong>
          </div>
        </div>
      </section>

      <v-row dense class="auth-admin__grid">
        <v-col cols="12" lg="8">
          <v-card flat class="auth-admin__card">
            <div class="auth-admin__card-head">
              <div>
                <div class="auth-admin__card-title">Diretório LDAP</div>
                <div class="auth-admin__card-copy">
                  O login continua centralizado na tela padrão do Semaphore. Quando o
                  LDAP está habilitado, o backend valida o usuário no diretório e
                  provisiona a conta externa automaticamente.
                </div>
              </div>

              <v-chip
                small
                label
                :color="form.ldap.enabled ? 'success' : 'default'"
                :outlined="!form.ldap.enabled"
              >
                {{ form.ldap.enabled ? 'Ativo' : 'Inativo' }}
              </v-chip>
            </div>

            <v-row dense class="mt-1">
              <v-col cols="12" md="6">
                <v-switch
                  v-model="form.ldap.enabled"
                  inset
                  class="mt-0"
                  label="Habilitar autenticação LDAP"
                />
              </v-col>

              <v-col cols="12" md="6">
                <v-switch
                  v-model="form.ldap.need_tls"
                  inset
                  class="mt-0"
                  label="Usar LDAPS / TLS direto"
                />
              </v-col>
            </v-row>

            <v-row dense>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.ldap.server"
                  outlined
                  dense
                  label="Servidor LDAP"
                  hint="Ex.: 172.25.200.100:389"
                  persistent-hint
                />
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.ldap.search_dn"
                  outlined
                  dense
                  label="Base DN / Search DN"
                  hint="Ex.: DC=clickip,DC=local"
                  persistent-hint
                />
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.ldap.bind_dn"
                  outlined
                  dense
                  label="Bind DN / usuário de consulta"
                  hint="Conta usada para pesquisar usuários no diretório"
                  persistent-hint
                />
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.ldap.bind_password"
                  outlined
                  dense
                  :type="showBindPassword ? 'text' : 'password'"
                  label="Senha do Bind DN"
                  :append-icon="showBindPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                  @click:append="showBindPassword = !showBindPassword"
                  :hint="bindPasswordHint"
                  persistent-hint
                />
              </v-col>
            </v-row>

            <v-checkbox
              v-if="form.ldap.has_bind_password"
              v-model="form.ldap.clear_bind_password"
              class="mt-0 pt-0"
              label="Remover a senha LDAP salva"
            />

            <v-text-field
              v-model="form.ldap.search_filter"
              outlined
              dense
              class="mt-2"
              label="Filtro de pesquisa"
              hint="Ex.: (&(objectClass=user)(sAMAccountName=%s))"
              persistent-hint
            />

            <div class="auth-admin__section-title">Mapeamento de atributos</div>
            <div class="auth-admin__section-copy">
              Esses campos definem como o Semaphore extrai login, nome, email e DN
              do usuário retornado pelo diretório.
            </div>

            <div class="auth-admin__mapping-tools">
              <v-btn
                small
                text
                color="primary"
                class="px-0"
                @click="applyActiveDirectoryPreset()"
              >
                Aplicar preset de Active Directory
              </v-btn>
            </div>

            <v-alert
              v-if="ldapMappingNeedsAttention"
              dense
              text
              color="warning"
              class="mt-2 mb-0"
            >
              O filtro indica Active Directory, mas o mapeamento ainda parece genérico.
              Para este cenário, use `sAMAccountName`, `displayName` e `distinguishedName`.
            </v-alert>

            <v-row dense class="mt-1">
              <v-col cols="12" md="3">
                <v-text-field
                  v-model="form.ldap.mappings.uid"
                  outlined
                  dense
                  label="UID / Login"
                />
              </v-col>

              <v-col cols="12" md="3">
                <v-text-field
                  v-model="form.ldap.mappings.cn"
                  outlined
                  dense
                  label="Nome completo"
                />
              </v-col>

              <v-col cols="12" md="3">
                <v-text-field
                  v-model="form.ldap.mappings.mail"
                  outlined
                  dense
                  label="E-mail"
                />
              </v-col>

              <v-col cols="12" md="3">
                <v-text-field
                  v-model="form.ldap.mappings.dn"
                  outlined
                  dense
                  label="DN"
                />
              </v-col>
            </v-row>

            <div class="auth-admin__section-title">Diagnóstico LDAP</div>
            <div class="auth-admin__section-copy">
              Use primeiro o teste de conexão. Depois valide um usuário real para
              confirmar o bind final e o mapeamento retornado.
            </div>

            <v-row dense class="mt-1">
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="test.login"
                  outlined
                  dense
                  label="Usuário de teste"
                  hint="Ex.: richard.barbosa ou samAccountName"
                  persistent-hint
                />
              </v-col>

              <v-col cols="12" md="6">
                <v-text-field
                  v-model="test.password"
                  outlined
                  dense
                  :type="showTestPassword ? 'text' : 'password'"
                  label="Senha do usuário de teste"
                  :append-icon="showTestPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
                  @click:append="showTestPassword = !showTestPassword"
                />
              </v-col>
            </v-row>

            <div class="auth-admin__actions">
              <v-btn
                outlined
                color="primary"
                :loading="testingConnection"
                @click="runLdapTest(false)"
              >
                Testar conexão LDAP
              </v-btn>

              <v-btn
                color="primary"
                :loading="testingUser"
                @click="runLdapTest(true)"
              >
                Testar autenticação do usuário
              </v-btn>
            </div>

            <v-alert
              v-if="testResult"
              dense
              outlined
              class="mt-5 mb-0"
              :type="testResult.type"
            >
              <div class="auth-admin__result-title">{{ testResult.message }}</div>

              <div v-if="testResult.user" class="auth-admin__result-grid">
                <div><strong>Nome:</strong> {{ testResult.user.name }}</div>
                <div><strong>Login:</strong> {{ testResult.user.username }}</div>
                <div><strong>E-mail:</strong> {{ testResult.user.email || 'não informado' }}</div>
              </div>
            </v-alert>
          </v-card>
        </v-col>

        <v-col cols="12" lg="4">
          <v-card flat class="auth-admin__card">
            <div class="auth-admin__card-title">Política global de TOTP</div>
            <div class="auth-admin__card-copy">
              O TOTP usa padrão compatível com Google Authenticator, Authy e apps
              equivalentes. A ativação continua sendo por usuário, mas a política
              global é controlada aqui.
            </div>

            <v-switch
              v-model="form.totp.enabled"
              inset
              class="mt-6"
              label="Habilitar TOTP na plataforma"
            />

            <v-switch
              v-model="form.totp.allow_recovery"
              inset
              class="mt-1"
              :disabled="!form.totp.enabled"
              label="Permitir código de recuperação"
            />

            <v-text-field
              v-model="form.totp.issuer"
              outlined
              dense
              class="mt-3"
              label="Issuer / nome do app"
              hint="Texto exibido no app autenticador"
              persistent-hint
            />

            <div class="auth-admin__note auth-admin__note--primary">
              Quando o TOTP está ativo, cada usuário precisa habilitar o segundo
              fator na própria conta em Segurança. O QR code e o recovery code
              aparecem nessa tela do usuário; eles não são enviados por email.
            </div>
          </v-card>

          <v-card flat class="auth-admin__card">
            <div class="auth-admin__card-title">Fluxo operacional</div>
            <div class="auth-admin__timeline">
              <div class="auth-admin__timeline-item">
                <span class="auth-admin__timeline-step">1</span>
                <div>
                  <strong>Salvar configuração</strong>
                  <p>A política é persistida no banco e aplicada em runtime.</p>
                </div>
              </div>

              <div class="auth-admin__timeline-item">
                <span class="auth-admin__timeline-step">2</span>
                <div>
                  <strong>Validar com usuário real</strong>
                  <p>O teste LDAP confirma bind, busca e mapeamento de atributos.</p>
                </div>
              </div>

              <div class="auth-admin__timeline-item">
                <span class="auth-admin__timeline-step">3</span>
                <div>
                  <strong>Promover para produção</strong>
                  <p>Depois do build, publique a tag e troque a imagem no compose.</p>
                </div>
              </div>
            </div>

            <div class="auth-admin__note">
              Para Active Directory, o filtro normalmente fica em
              <code>(&amp;(objectClass=user)(sAMAccountName=%s))</code>.
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import EventBus from '@/event-bus';
import { getErrorMessage } from '@/lib/error';

const defaultSettings = () => ({
  ldap: {
    enabled: false,
    server: '',
    need_tls: false,
    bind_dn: '',
    bind_password: '',
    clear_bind_password: false,
    has_bind_password: false,
    search_dn: '',
    search_filter: '(&(objectClass=user)(sAMAccountName=%s))',
    mappings: {
      dn: 'distinguishedName',
      uid: 'sAMAccountName',
      cn: 'displayName',
      mail: 'mail',
    },
  },
  totp: {
    enabled: false,
    allow_recovery: false,
    issuer: 'I-Mais Semaphore',
  },
});

export default {
  data() {
    return {
      loading: true,
      loadError: null,
      saving: false,
      testingConnection: false,
      testingUser: false,
      showBindPassword: false,
      showTestPassword: false,
      form: defaultSettings(),
      test: {
        login: '',
        password: '',
      },
      testResult: null,
    };
  },

  async created() {
    await this.loadSettings();
  },

  watch: {
    'form.ldap.bind_password': function onBindPasswordChange(value) {
      if (value) {
        this.form.ldap.clear_bind_password = false;
      }
    },
  },

  computed: {
    bindPasswordHint() {
      if (this.form.ldap.clear_bind_password) {
        return 'A senha salva será removida ao salvar.';
      }

      if (this.form.ldap.has_bind_password) {
        return 'Deixe em branco para manter a senha LDAP já armazenada.';
      }

      return 'Informe a senha apenas se o Bind DN exigir autenticação.';
    },

    ldapMappingNeedsAttention() {
      const filter = (this.form.ldap.search_filter || '').toLowerCase();
      if (!filter.includes('samaccountname')) {
        return false;
      }

      const mappings = this.form.ldap.mappings || {};
      return mappings.uid === 'uid'
        || mappings.cn === 'cn'
        || mappings.dn === 'dn';
    },
  },

  methods: {
    applyActiveDirectoryPreset() {
      this.form.ldap.search_filter = '(&(objectClass=user)(sAMAccountName=%s))';
      this.form.ldap.mappings = {
        dn: 'distinguishedName',
        uid: 'sAMAccountName',
        cn: 'displayName',
        mail: 'mail',
      };
    },

    getPayload() {
      return {
        ldap: {
          enabled: this.form.ldap.enabled,
          server: this.form.ldap.server,
          need_tls: this.form.ldap.need_tls,
          bind_dn: this.form.ldap.bind_dn,
          bind_password: this.form.ldap.bind_password,
          clear_bind_password: this.form.ldap.clear_bind_password,
          search_dn: this.form.ldap.search_dn,
          search_filter: this.form.ldap.search_filter,
          mappings: {
            dn: this.form.ldap.mappings.dn,
            uid: this.form.ldap.mappings.uid,
            cn: this.form.ldap.mappings.cn,
            mail: this.form.ldap.mappings.mail,
          },
        },
        totp: {
          enabled: this.form.totp.enabled,
          allow_recovery: this.form.totp.allow_recovery,
          issuer: this.form.totp.issuer,
        },
      };
    },

    applySettings(data) {
      const defaults = defaultSettings();

      this.form = {
        ldap: {
          enabled: data.ldap.enabled,
          server: data.ldap.server || '',
          need_tls: data.ldap.need_tls,
          bind_dn: data.ldap.bind_dn || '',
          bind_password: '',
          clear_bind_password: false,
          has_bind_password: data.ldap.has_bind_password,
          search_dn: data.ldap.search_dn || '',
          search_filter: data.ldap.search_filter || defaults.ldap.search_filter,
          mappings: {
            dn: (data.ldap.mappings || {}).dn || defaults.ldap.mappings.dn,
            uid: (data.ldap.mappings || {}).uid || defaults.ldap.mappings.uid,
            cn: (data.ldap.mappings || {}).cn || defaults.ldap.mappings.cn,
            mail: (data.ldap.mappings || {}).mail || defaults.ldap.mappings.mail,
          },
        },
        totp: {
          enabled: data.totp.enabled,
          allow_recovery: data.totp.allow_recovery,
          issuer: data.totp.issuer || defaults.totp.issuer,
        },
      };
    },

    async loadSettings() {
      this.loading = true;
      this.loadError = null;

      try {
        const response = await axios({
          method: 'get',
          url: '/api/auth/settings',
          responseType: 'json',
        });

        this.applySettings(response.data);
      } catch (error) {
        this.loadError = error.response && error.response.status === 403
          ? 'Acesso restrito a administradores.'
          : getErrorMessage(error);
      } finally {
        this.loading = false;
      }
    },

    async saveSettings() {
      this.saving = true;

      try {
        const response = await axios({
          method: 'post',
          url: '/api/auth/settings',
          responseType: 'json',
          data: this.getPayload(),
        });

        this.applySettings(response.data);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Configurações de autenticação salvas com sucesso.',
        });
      } catch (error) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(error),
        });
      } finally {
        this.saving = false;
      }
    },

    async runLdapTest(withCredentials) {
      if (withCredentials && (!this.test.login || !this.test.password)) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: 'Informe usuário e senha para validar a autenticação LDAP.',
        });
        return;
      }

      this.testResult = null;

      if (withCredentials) {
        this.testingUser = true;
      } else {
        this.testingConnection = true;
      }

      try {
        const response = await axios({
          method: 'post',
          url: '/api/auth/settings/ldap/test',
          responseType: 'json',
          data: {
            ldap: this.getPayload().ldap,
            login: withCredentials ? this.test.login : '',
            password: withCredentials ? this.test.password : '',
          },
        });

        this.testResult = {
          type: 'success',
          message: response.data.message,
          user: response.data.user,
        };
      } catch (error) {
        this.testResult = {
          type: 'error',
          message: getErrorMessage(error),
          user: null,
        };
      } finally {
        this.testingConnection = false;
        this.testingUser = false;
      }
    },

    returnToProjects() {
      EventBus.$emit('i-open-last-project');
    },
  },
};
</script>

<style lang="scss" scoped>
.auth-admin {
  min-height: calc(100dvh - 32px);
  padding: 20px 24px 32px;
}

.auth-admin__loading {
  align-items: center;
  display: flex;
  justify-content: center;
  min-height: calc(100dvh - 96px);
}

.auth-admin__error-card {
  border: 1px solid rgba(37, 99, 235, 0.12);
  border-radius: 24px;
  max-width: 520px;
  padding: 28px;
  width: 100%;
}

.auth-admin__error-title {
  color: #10243b;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.auth-admin__error-copy {
  color: #5f7389;
  line-height: 1.7;
  margin-top: 10px;
}

.auth-admin__toolbar {
  background: transparent !important;
  margin-bottom: 12px;
  padding-left: 0;
  padding-right: 0;
}

.auth-admin__eyebrow {
  color: #4d6480;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.auth-admin__title {
  color: #10243b;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: -0.05em;
  line-height: 1;
  margin-top: 6px;
}

.auth-admin__save-btn {
  border-radius: 14px !important;
  min-width: 176px;
}

.auth-admin__hero {
  background:
    radial-gradient(circle at top right, rgba(84, 182, 255, 0.24), transparent 30%),
    linear-gradient(135deg, #eff6ff 0%, #f8fbff 42%, #edf7ff 100%);
  border: 1px solid rgba(37, 99, 235, 0.12);
  border-radius: 28px;
  display: grid;
  gap: 18px;
  grid-template-columns: minmax(0, 1.4fr) minmax(260px, 0.7fr);
  margin-bottom: 18px;
  padding: 24px 28px;
}

.auth-admin__hero-kicker {
  color: #0f5f78;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.auth-admin__hero-title {
  color: #12263f;
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -0.05em;
  line-height: 1.02;
  margin-top: 14px;
}

.auth-admin__hero-copy {
  color: #536b84;
  font-size: 15px;
  line-height: 1.8;
  margin: 14px 0 0;
  max-width: 760px;
}

.auth-admin__hero-badges {
  align-content: center;
  display: grid;
  gap: 12px;
}

.auth-admin__hero-badge {
  background: rgba(255, 255, 255, 0.84);
  border: 1px solid rgba(15, 118, 110, 0.12);
  border-radius: 22px;
  padding: 18px 20px;
}

.auth-admin__hero-badge-label {
  color: #6b7f96;
  display: block;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.12em;
  margin-bottom: 8px;
  text-transform: uppercase;
}

.auth-admin__hero-badge strong {
  color: #10243b;
  font-size: 20px;
  letter-spacing: -0.03em;
}

.auth-admin__grid {
  margin-top: 0 !important;
}

.auth-admin__card {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 26px;
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.05);
  height: 100%;
  padding: 22px;
}

.auth-admin__card + .auth-admin__card {
  margin-top: 12px;
}

.auth-admin__card-head {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
  margin-bottom: 18px;
}

.auth-admin__card-title {
  color: #10243b;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.1;
}

.auth-admin__card-copy {
  color: #61768d;
  line-height: 1.7;
  margin-top: 10px;
}

.auth-admin__section-title {
  color: #17324f;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: -0.02em;
  margin-top: 18px;
}

.auth-admin__section-copy {
  color: #63778e;
  line-height: 1.7;
  margin-top: 8px;
}

.auth-admin__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
}

.auth-admin__result-title {
  font-weight: 700;
}

.auth-admin__result-grid {
  display: grid;
  gap: 6px;
  margin-top: 10px;
}

.auth-admin__note {
  background: linear-gradient(135deg, rgba(15, 118, 110, 0.06), rgba(59, 130, 246, 0.08));
  border: 1px solid rgba(15, 118, 110, 0.12);
  border-radius: 18px;
  color: #46627a;
  line-height: 1.7;
  margin-top: 18px;
  padding: 16px 18px;
}

.auth-admin__note--primary {
  margin-top: 22px;
}

.auth-admin__timeline {
  display: grid;
  gap: 16px;
  margin-top: 20px;
}

.auth-admin__timeline-item {
  align-items: flex-start;
  display: flex;
  gap: 14px;
}

.auth-admin__timeline-item p {
  color: #63778e;
  line-height: 1.65;
  margin: 6px 0 0;
}

.auth-admin__timeline-step {
  align-items: center;
  background: linear-gradient(135deg, #17324f, #2f74b7);
  border-radius: 14px;
  color: #ffffff;
  display: inline-flex;
  font-size: 13px;
  font-weight: 700;
  height: 30px;
  justify-content: center;
  min-width: 30px;
}

@media (max-width: 1264px) {
  .auth-admin__hero {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 960px) {
  .auth-admin {
    padding: 16px 16px 28px;
  }

  .auth-admin__toolbar {
    align-items: flex-start;
    flex-wrap: wrap;
    gap: 12px;
  }

  .auth-admin__save-btn {
    width: 100%;
  }

  .auth-admin__hero {
    padding: 20px;
  }

  .auth-admin__hero-title {
    font-size: 30px;
  }

  .auth-admin__title {
    font-size: 28px;
  }
}
</style>
