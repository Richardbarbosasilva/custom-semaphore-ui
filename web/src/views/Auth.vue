<template>
  <div class="auth">
    <v-dialog v-model="loginHelpDialog" max-width="600">
      <v-card>
        <v-card-title>
          {{ $t('howToFixSigninIssues') }}
          <v-spacer></v-spacer>
          <v-btn icon @click="loginHelpDialog = false">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>
        <v-card-text>
          <p class="text-body-1">
            {{ $t('firstlyYouNeedAccessToTheServerWhereSemaphoreRunni') }}
          </p>
          <p class="text-body-1">
            {{ $t('executeTheFollowingCommandOnTheServerToSeeExisting') }}
          </p>
          <v-alert
            dense
            text
            color="info"
            style="font-family: monospace;"
          >
            {{ $t('semaphoreUserList') }}
          </v-alert>
          <p class="text-body-1">
            {{ $t('youCanChangePasswordOfExistingUser') }}
          </p>
          <v-alert
            dense
            text
            color="info"
            style="font-family: monospace;"
          >
            {{
              $t('semaphoreUserChangebyloginLoginUser123Password', {
                makePasswordExample:
                  makePasswordExample()
              })
            }}
          </v-alert>
          <p class="text-body-1">
            {{ $t('orCreateNewAdminUser') }}
          </p>
          <v-alert
            dense
            text
            color="info"
            style="font-family: monospace;"
          >
            semaphore user add --admin --login user123 --name User123
            --email user123@example.com --password {{ makePasswordExample() }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer/>
          <v-btn
            color="blue darken-1"
            text
            @click="loginHelpDialog = false"
          >
            {{ $t('close2') }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <div class="auth__layout">
      <section class="auth__hero" :style="authHeroStyle">
        <div class="auth__hero-overlay"></div>
        <div class="auth__hero-content">
          <img :src="authHeroLogo" alt="iMais" class="auth__hero-logo" />

          <div class="auth__hero-copy">
            <div class="auth__hero-eyebrow">{{ $t('auth_eyebrow') }}</div>
            <h1 class="auth__hero-title">{{ $t('auth_headline') }}</h1>
            <p class="auth__hero-subtitle">{{ $t('auth_subtitle') }}</p>
          </div>
        </div>
      </section>

      <aside class="auth__panel">
        <div class="auth__panel-inner">
          <div class="auth__panel-brand">
            <img :src="authPanelLogo" alt="iMais" class="auth__panel-logo" />
          </div>

          <v-card class="auth__card" flat>
            <v-card-text class="pa-0">
              <v-form
                @submit.prevent
                ref="signInForm"
                lazy-validation
                v-model="signInFormValid"
                class="auth__form"
              >
                <h2 v-if="screen === 'verification'" class="auth__form-title">
                  Two-step verification
                </h2>

                <h2 v-else-if="screen === 'recovery'" class="auth__form-title">
                  Account recovery
                </h2>

                <h2 v-else class="auth__form-title">
                  Enter to your account
                </h2>

                <div class="auth__form-subtitle" v-if="screen == null">
                  {{ $t('signIn') }} {{ $t('ansibleSemaphore') }}
                </div>

                <v-alert
                  :value="signInError != null"
                  color="error"
                  class="auth__alert"
                >{{ signInError }}
                </v-alert>

                <div v-if="screen === 'verification'">

                  <div v-if="verificationMethod === 'totp'" class="text-center mb-4">
                    Open the two-step verification app on your mobile device to
                    get your verification code.
                  </div>

                  <div
                    v-else-if="isPortal && verificationMethod === 'email'"
                    class="text-center mb-4"
                  >
                    Check your email for the verification code we just sent you.
                  </div>

                  <v-otp-input
                    v-model="verificationCode"
                    length="6"
                    @finish="verify()"
                  ></v-otp-input>

                  <v-divider class="my-6" />

                  <div class="text-center">
                    <a @click="signOut()" class="mr-6">{{ $t('Return to login') }}</a>
                    <a
                      v-if="verificationMethod === 'totp'
                        && authMethods.totp
                        && authMethods.totp.allow_recovery"
                      @click="screen = 'recovery'"
                    >
                      {{ $t('Use recovery code') }}
                    </a>

                    <v-btn
                      :width="200"
                      small
                      :disabled="verificationEmailSending"
                      color="primary"
                      v-if="isPortal && verificationMethod === 'email'"
                      @click="resendEmailVerification()"
                    >
                      {{
                        verificationEmailSending
                          ? $t('Email sending...')
                          : $t('Resend code to email')
                      }}
                    </v-btn>
                  </div>
                </div>

                <div v-else-if="screen === 'recovery'">
                  <div class="text-center mb-2">
                    Use your recovery code to regain access to your account.
                  </div>

                  <v-text-field
                    class="auth__field mt-6"
                    outlined
                    v-model="recoveryCode"
                    @keyup.enter.native="signIn"
                    :label="$t('Recovery code')"
                    :rules="[v => !!v || $t('recoveryCode_required')]"
                    required
                  />

                  <div>
                    <v-btn
                      class="auth__primary-btn"
                      color="primary"
                      @click="recovery()"
                      block
                      rounded
                    >
                      Send
                    </v-btn>
                  </div>

                  <div class="text-center pt-6">
                    <a @click="screen = 'verification'">{{ $t('Return to verification') }}</a>
                  </div>

                </div>

                <div v-else>

                  <div v-if="loginWithPassword">
                    <v-text-field
                      v-model="username"
                      outlined
                      class="auth__field"
                      v-bind:label='$t("username")'
                      :rules="[v => !!v || $t('username_required')]"
                      required
                      :disabled="signInProcess"
                      id="auth-username"
                      data-testid="auth-username"
                    ></v-text-field>

                    <v-text-field
                      v-model="password"
                      outlined
                      class="auth__field"
                      :label="$t('password')"
                      :rules="[v => !!v || $t('password_required')]"
                      type="password"
                      required
                      :disabled="signInProcess"
                      @keyup.enter.native="signIn"
                      id="auth-password"
                      data-testid="auth-password"
                    ></v-text-field>

                    <v-btn
                      large
                      color="primary"
                      @click="signIn"
                      :disabled="signInProcess"
                      class="auth__primary-btn"
                      block
                      rounded
                      data-testid="auth-signin"
                    >
                      {{ $t('signIn') }}
                    </v-btn>

                  </div>

                  <div v-else-if="isPortal">
                    <v-text-field
                      v-model="email"
                      :label="$t('Email')"
                      :rules="[v => !!v || $t('email_required')]"
                      type="email"
                      required
                      outlined
                      class="auth__field"
                      :disabled="signInProcess"
                      @keyup.enter.native="signInWithEmail"
                      data-testid="auth-password"
                    ></v-text-field>

                    <v-btn
                      large
                      color="primary"
                      @click="signInWithEmail"
                      :disabled="signInProcess"
                      class="auth__primary-btn"
                      block
                      rounded
                      data-testid="auth-signin-with-eamil"
                    >
                      <v-icon
                        left
                        dark
                      >
                        mdi-email
                      </v-icon>

                      {{ $t('Continue with Email') }}
                    </v-btn>
                  </div>

                  <div
                    class="auth__divider"
                    v-if="(loginWithPassword || isPortal) && oidcProviders.length > 0"
                  >or</div>

                  <v-btn
                    large
                    v-for="provider in oidcProviders"
                    :color="provider.color || 'secondary'"
                    dark
                    class="auth__provider-btn"
                    @click="oidcSignIn(provider.id)"
                    block
                    :key="provider.id"
                    rounded
                  >
                    <v-icon
                      left
                      dark
                      v-if="provider.icon"
                    >
                      mdi-{{ provider.icon }}
                    </v-icon>

                    {{ provider.name }}
                  </v-btn>
                </div>
              </v-form>
            </v-card-text>
          </v-card>
        </div>
      </aside>
    </div>
  </div>
</template>
<style lang="scss">
.auth__divider {
  margin-bottom: 8px;
  margin-top: 18px;

  display: flex;
  &:before, &:after {
    margin-top: 10px;
    width: 100%;
    content: "";
    border-top: 1px solid rgba(128, 128, 128, 0.51);
  }

  &:before {
    margin-right: 10px;
  }

  &:after {
    margin-left: 10px;
  }
}

.auth {
  height: 100dvh;
  overflow: hidden;
}

.auth__layout {
  display: flex;
  height: 100%;
}

.auth__hero {
  background-position: center;
  background-size: cover;
  color: #ffffff;
  flex: 1 1 auto;
  overflow: hidden;
  position: relative;
}

.auth__hero-overlay {
  backdrop-filter: blur(1px);
  background:
    linear-gradient(135deg, rgba(7, 12, 26, 0.94), rgba(8, 30, 48, 0.66)),
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.22), transparent 32%);
  inset: 0;
  position: absolute;
}

.auth__hero-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  justify-content: space-between;
  padding: 42px 56px;
  position: relative;
  z-index: 1;
}

.auth__hero-logo {
  max-width: 180px;
}

.auth__hero-copy {
  margin-bottom: 8vh;
  max-width: 560px;
}

.auth__hero-eyebrow {
  color: rgba(233, 241, 255, 0.72);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.auth__hero-title {
  font-size: 52px;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 0.98;
  margin-top: 16px;
}

.auth__hero-subtitle {
  color: rgba(232, 240, 255, 0.78);
  font-size: 18px;
  line-height: 1.7;
  margin-top: 20px;
}

.v-application .primary {
    background-color: #183149 !important;
    border-color: #1976d2 !important;
}

.auth__panel {
    align-items: center;
    background: linear-gradient(32deg, #ffffff 0%, #a9caff 100%);
    display: flex;
    flex: 0 0 420px;
    justify-content: center;
    overflow-y: auto;
    padding: 24px;
}

.auth__panel-inner {
  max-width: 340px;
  width: 100%;
}

.auth__panel-brand {
  display: flex;
  justify-content: center;
  margin-bottom: 28px;
}

.auth__panel-logo {
  max-width: 230px;
}

.auth__card {
  background: transparent !important;
}

.v-application .primary {
    background-color: #1b4166 !important;
    border-color: #1976d2 !important;
}

.auth__form-title {
  color: #0f172a;
  font-size: 34px;
  font-weight: 800;
  letter-spacing: -0.04em;
  line-height: 1.02;
  text-align: center;
}

.auth__form-subtitle {
  color: #64748b;
  font-size: 14px;
  margin-bottom: 28px;
  margin-top: 10px;
}

.auth__alert {
  margin-bottom: 20px;
}

.auth__field {
  margin-bottom: 10px;
}

.auth__field .v-input__slot {
  background: #ffffff !important;
  border-radius: 14px !important;
}

.auth__primary-btn,
.auth__provider-btn {
  min-height: 52px !important;
}

.auth__provider-btn {
  margin-top: 12px;
}

@media (max-width: 1200px) {
  .auth__hero-title {
    font-size: 42px;
  }
}

@media (max-width: 960px) {
  .auth__layout {
    flex-direction: column;
  }

  .auth__hero {
    min-height: 34vh;
  }

  .auth__hero-content {
    justify-content: flex-end;
    padding: 28px 24px;
  }

  .auth__hero-logo {
    max-width: 140px;
  }

  .auth__hero-copy {
    margin-bottom: 0;
    max-width: 100%;
  }

  .auth__hero-title {
    font-size: 34px;
  }

  .auth__hero-subtitle {
    font-size: 15px;
    line-height: 1.6;
  }

  .auth__panel {
    flex: 1 1 auto;
    padding: 22px 18px 28px;
  }
}
</style>
<script>
import axios from 'axios';
import { getErrorMessage } from '@/lib/error';
import EventBus from '@/event-bus';
import authHeroLogo from '@/assets/i_mais_white.png';
import authPanelLogo from '@/assets/i_mais_azul.png';
import authHeroMedia from '@/assets/wallpaper_semaphore_new.gif';

export default {
  data() {
    return {
      signInFormValid: false,
      signInError: null,
      signInProcess: false,

      password: null,
      username: null,

      email: null,

      loginHelpDialog: null,

      oidcProviders: [],
      loginWithPassword: null,
      authMethods: {},

      screen: null,

      verificationCode: null,
      verificationMethod: null,
      recoveryCode: null,
      verificationEmailSending: false,
      authHeroLogo,
      authPanelLogo,
      authHeroMedia,

    };
  },

  async created() {
    const { status, verificationMethod } = await this.getAuthenticationStatus();

    switch (status) {
      case 'authenticated':
        this.redirectAfterLogin();
        break;
      case 'unauthenticated':
        await this.loadLoginData();
        break;
      case 'unverified':
        this.screen = 'verification';
        this.verificationMethod = verificationMethod;
        await this.loadLoginData();
        break;
      default:
        throw new Error(`Unknown authentication status: ${status}`);
    }
  },

  computed: {
    isPortal() {
      return process.env.VUE_APP_BUILD_TYPE === 'pro_portal';
    },

    authHeroStyle() {
      return {
        backgroundImage: `url(${this.authHeroMedia})`,
      };
    },
  },

  methods: {
    async resendEmailVerification() {
      if (this.verificationEmailSending) {
        return;
      }

      this.verificationEmailSending = true;
      try {
        (await axios({
          method: 'post',
          url: '/api/auth/login/email/resend',
          responseType: 'json',
        }));
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Verification email sent successfully.',
        });
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(e),
        });
      } finally {
        this.verificationEmailSending = false;
      }
    },

    async loadLoginData() {
      await axios({
        method: 'get',
        url: '/api/auth/login',
        responseType: 'json',
      }).then((resp) => {
        this.oidcProviders = resp.data.oidc_providers;
        this.loginWithPassword = resp.data.login_with_password;
        this.authMethods = resp.data.auth_methods || {};
      });
    },

    async recovery() {
      this.signInProcess = true;

      try {
        await axios({
          method: 'post',
          url: '/api/auth/recovery',
          responseType: 'json',
          data: {
            recovery_code: this.recoveryCode,
          },
        });

        const { location } = document;
        document.location = location;
      } catch (e) {
        this.signInError = getErrorMessage(e);
      } finally {
        this.signInProcess = false;
      }
    },

    async signOut() {
      try {
        (await axios({
          method: 'post',
          url: '/api/auth/logout',
          responseType: 'json',
        }));

        const { location } = document;
        document.location = location;
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(e),
        });
      }
    },

    makePasswordExample() {
      let pwd = '';
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
      const charactersLength = characters.length;
      for (let i = 0; i < 10; i += 1) {
        pwd += characters.charAt(Math.floor(Math.random() * charactersLength));
      }
      return pwd;
    },

    async getAuthenticationStatus() {
      try {
        await axios({
          method: 'get',
          url: '/api/user',
          responseType: 'json',
        });
      } catch (err) {
        if (err.response.status === 401) {
          switch (err.response.data.error) {
            case 'TOTP_REQUIRED':
              return {
                status: 'unverified',
                verificationMethod: 'totp',
              };
            case 'EMAIL_OTP_REQUIRED':
              return {
                status: 'unverified',
                verificationMethod: 'email',
              };
            default:
              return { status: 'unauthenticated' };
          }
        }
        throw err;
      }

      return { status: 'authenticated' };
    },

    async verify() {
      this.signInError = null;

      if (!this.$refs.signInForm.validate()) {
        return;
      }

      this.signInProcess = true;

      try {
        await axios({
          method: 'post',
          url: '/api/auth/verify',
          responseType: 'json',
          data: {
            passcode: this.verificationCode,
          },
        });

        this.redirectAfterLogin();
      } catch (err) {
        this.signInError = getErrorMessage(err);
      } finally {
        this.signInProcess = false;
      }
    },

    async signInWithEmail() {
      this.signInError = null;

      if (!this.$refs.signInForm.validate()) {
        return;
      }

      this.signInProcess = true;
      try {
        await axios({
          method: 'post',
          url: '/api/auth/login/email',
          responseType: 'json',
          data: {
            email: this.email,
          },
        });

        this.redirectAfterLogin();
      } catch (err) {
        if (err.response.status === 401) {
          this.signInError = this.$t('incorrectEmail');
        } else {
          this.signInError = getErrorMessage(err);
        }
      } finally {
        this.signInProcess = false;
      }
    },

    async signIn() {
      this.signInError = null;

      if (!this.$refs.signInForm.validate()) {
        return;
      }

      this.signInProcess = true;
      try {
        await axios({
          method: 'post',
          url: '/api/auth/login',
          responseType: 'json',
          data: {
            auth: this.username,
            password: this.password,
          },
        });

        this.redirectAfterLogin();
        // document.location = document.baseURI + window.location.search;
      } catch (err) {
        if (err.response.status === 401) {
          this.signInError = this.$t('incorrectUsrPwd');
        } else {
          this.signInError = getErrorMessage(err);
        }
      } finally {
        this.signInProcess = false;
      }
    },

    async oidcSignIn(provider) {
      const params = new URLSearchParams();
      const returnTo = this.$route.query.return;
      if (returnTo) {
        params.set('return', returnTo);
      } else if (this.$route.query.new_project === 'premium') {
        params.set('return', '/project/premium');
      }
      const qs = params.toString();
      const suffix = qs ? `?${qs}` : '';
      document.location = `${document.baseURI}api/auth/oidc/${provider}/login${suffix}`;
    },

    redirectAfterLogin() {
      const redirectTo = this.$route.query.return;
      let baseURI = document.baseURI;

      if (redirectTo) {
        if (baseURI.endsWith('/')) {
          baseURI = baseURI.substring(0, baseURI.length - 1);
        }

        document.location = baseURI + redirectTo;

        return;
      }

      document.location = document.baseURI + window.location.search;
    },
  },
};
</script>
