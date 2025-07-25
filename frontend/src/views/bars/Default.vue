<template>
  <header v-if="!isOnlyOffice" :class="['flexbar', { 'dark-mode-header': isDarkMode }]">
    <action
      v-if="!isShare && !(disableNavButtons && isListingView)"
      icon="close_back"
      :label="$t('buttons.close')"
      :disabled="isDisabledMultiAction"
      @action="multiAction"
    />
    <search v-if="showSearch" />
    <title v-else-if="isSettings" class="topTitle">{{ $t("sidebar.settings") }}</title>
    <title v-else class="topTitle">{{ req.name }}</title>
    <action
      v-if="isListingView && !disableNavButtons"
      class="menu-button"
      :icon="viewIcon"
      :label="$t('buttons.switchView')"
      @action="switchView"
      :disabled="isDisabled"
    />
    <action
      v-else-if="!isShare && !isListingView && !showQuickSave"
      :icon="iconName"
      :disabled="noItems"
      @click="toggleOverflow"
    />
    <action
      v-else-if="showQuickSave"
      id="save-button"
      icon="save"
      :label="$t('buttons.save')"
      @action="save()"
    />
  </header>
</template>

<script>
import router from "@/router";
import buttons from "@/utils/buttons";
import { notify } from "@/notify";
import { eventBus } from "@/store/eventBus";
import { getters, state, mutations } from "@/store";
import Action from "@/components/Action.vue";
import Search from "@/components/Search.vue";
import { disableNavButtons } from "@/utils/constants";

export default {
  name: "UnifiedHeader",
  components: {
    Action,
    Search,
  },
  data() {
    return {
      viewModes: ["list", "compact", "normal", "gallery"],
    };
  },
  computed: {
    showQuickSave() {
      console.log("showQuickSave", getters.currentView());
      if (getters.currentView() != "editor" || !state.user.permissions.modify) {
        return false;
      }
      return state.user.editorQuickSave;
    },
    disableNavButtons() {
      return disableNavButtons && !state.user.permissions.admin;
    },
    isOnlyOffice() {
      return getters.currentView() === "onlyOfficeEditor";
    },
    isListingView() {
      return getters.currentView() == "listingView";
    },
    iconName() {
      return getters.currentPromptName() === "OverflowMenu"
        ? "keyboard_arrow_up"
        : "more_vert";
    },
    viewIcon() {
      const icons = {
        list: "view_list",
        compact: "table_rows_narrow",
        normal: "view_module",
        gallery: "grid_view",
      };
      return icons[state.user.viewMode] || "grid_view";
    },
    isShare() {
      return getters.currentView() == "share";
    },
    noItems() {
      return !this.showEdit && !this.showSave && !this.showDelete;
    },
    showEdit() {
      return window.location.hash != "#edit" && state.user.permissions.modify;
    },
    showDelete() {
      return state.user.permissions.modify && getters.currentView() == "preview";
    },
    showSave() {
      return getters.currentView() == "editor" && state.user.permissions.modify;
    },
    showSearch() {
      return getters.isLoggedIn() && getters.currentView() === "listingView";
    },
    isDisabled() {
      return state.isSearchActive || getters.currentPromptName() != null;
    },
    isDisabledMultiAction() {
      return this.isDisabled || (getters.isStickySidebar() && state.multiButtonState === "menu");
    },
    showSwitchView() {
      return getters.currentView() === "listingView";
    },
    showSidebarToggle() {
      return getters.currentView() === "listingView";
    },
    req() {
      return state.req;
    },
    isDarkMode() {
      return getters.isDarkMode();
    },
    isSettings() {
      return getters.isSettings();
    },
  },
  methods: {
    async save() {
      const button = "save";
      buttons.loading("save");
      try {
        eventBus.emit("handleEditorValueRequest", "data");
        buttons.success(button);
        notify.showSuccess("File Saved!");
      } catch (e) {
        buttons.done(button);
        notify.showError("Error saving file: ", e);
      }
    },
    toggleOverflow() {
      if (getters.currentPromptName() === "OverflowMenu") {
        mutations.closeHovers();
      } else {
        mutations.showHover({ name: "OverflowMenu" });
      }
    },
    switchView() {
      mutations.closeHovers();
      const index = this.viewModes.indexOf(state.user.viewMode);
      const next = (index + 1) % this.viewModes.length;
      mutations.updateCurrentUser({ viewMode: this.viewModes[next] });
    },
    multiAction() {
      const listingView = getters.currentView();
      if (listingView == "listingView") {
        mutations.toggleSidebar();
      } else {
        mutations.closeHovers();
        if (listingView === "settings") {
          router.push({ path: "/files" });
          return;
        }
        mutations.replaceRequest({});
        router.go(-1);
      }
    },
  },
};
</script>
