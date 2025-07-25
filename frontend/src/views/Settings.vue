<template>
  <div class="dashboard" style="padding-bottom: 30vh">
    <div v-if="isRootSettings && !userPage" class="settings-views">
      <div
        v-for="setting in settings"
        :key="setting.id + '-main'"
        :id="setting.id + '-main'"
        @click="handleClick($event, setting.id + '-main')"
      >
        <!-- Dynamically render the component based on the setting -->
        <component v-if="shouldShow(setting)" :is="setting.component"></component>
      </div>
    </div>
    <div v-else class="settings-views">
      <div class="active">
        <UserSettings />
      </div>
    </div>

    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
  </div>
</template>

<script>
import { state, getters, mutations } from "@/store";
import { settings } from "@/utils/constants";
import GlobalSettings from "@/views/settings/Global.vue";
import ProfileSettings from "@/views/settings/Profile.vue";
import SharesSettings from "@/views/settings/Shares.vue";
import UserManagement from "@/views/settings/Users.vue";
import UserSettings from "@/views/settings/User.vue";
import FileLoading from "@/views/settings/FileLoading.vue";
import ApiKeys from "@/views/settings/Api.vue";
export default {
  name: "settings",
  components: {
    UserManagement,
    UserSettings,
    GlobalSettings,
    ProfileSettings,
    SharesSettings,
    ApiKeys,
    FileLoading,
  },
  data() {
    return {
      settings, // Initialize the settings array in data
    };
  },
  computed: {
    isRootSettings() {
      return getters.currentView() == "settings";
    },
    userPage() {
      return getters.routePath().startsWith(`/settings/users/`);
    },
    loading() {
      return getters.isLoading();
    },
    user() {
      return state.user;
    },
    currentHash() {
      return getters.currentHash();
    },
  },
  mounted() {
    mutations.closeHovers();
    mutations.setSearch(false);
  },
  methods: {
    shouldShow(setting) {
      const perm = setting?.permissions || {};
      return Object.keys(perm).every((key) => state.user.permissions[key]);
    },
    setView(view) {
      if (state.activeSettingsView === view) return;
      mutations.setActiveSettingsView(view);
    },
    handleClick(event, view) {
      // Allow propagation if the click is on a link or a child element with default behavior
      const target = event.target.closest("a, router-link");
      if (target) return; // Let the browser/router handle the navigation
      this.setView(view); // Call the setView method for other clicks
    },
  },
};
</script>

<style>

.form-group {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
}

.form-button {
  border-top-left-radius: 0 !important;
  border-bottom-left-radius: 0 !important;
  height: auto !important
}

.invalid-form {
  border-color: red !important;
}

.flat-right {
  border-top-right-radius: 0 !important;
  border-bottom-right-radius: 0 !important;
}
.flat-left {
  border-top-left-radius: 0 !important;
  border-bottom-left-radius: 0 !important;
}

.dashboard {
  display: flex;
  flex-direction: column;
  height: 100%;
  align-items: center;
}

.settings-views {
  max-width: 1000px;
  width: 100%;
}


.settings-views .card {
  border-style: solid;
  opacity: 1;
}

.settings-items > .item {
  padding: 1em;
  border-radius: 1em;
}

.settings-items > .item:hover {
  background-color: var(--surfaceSecondary);
}
</style>
