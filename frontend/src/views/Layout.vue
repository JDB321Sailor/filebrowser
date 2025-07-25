<template>
  <div>
    <div
      v-show="showOverlay"
      @contextmenu.prevent="onOverlayRightClick"
      @click="resetPrompts"
      class="overlay"
    ></div>
    <div v-if="progress" class="progress">
      <div v-bind:style="{ width: this.progress + '%' }"></div>
    </div>
    <defaultBar :class="{ 'dark-mode-header': isDarkMode }"></defaultBar>
    <sidebar></sidebar>
    <Scrollbar
      id="main"
      :class="{
        'dark-mode': isDarkMode,
        moveWithSidebar: moveWithSidebar,
        'remove-padding-top': isOnlyOffice,
        'main-padding': showPadding,
        scrollable: scrollable,
      }"
    >
      <router-view />
    </Scrollbar>
    <prompts :class="{ 'dark-mode': isDarkMode }"></prompts>
  </div>
  <Notifications />
  <ContextMenu></ContextMenu>
  <Tooltip />
</template>

<script>
import defaultBar from "./bars/Default.vue";
import Prompts from "@/components/prompts/Prompts.vue";
import Sidebar from "@/components/sidebar/Sidebar.vue";
import ContextMenu from "@/components/ContextMenu.vue";
import Notifications from "@/components/Notifications.vue";
import Scrollbar from "@/components/files/Scrollbar.vue";
import Tooltip from "@/components/Tooltip.vue";
import { filesApi } from "@/api";
import { state, getters, mutations } from "@/store";
import { events } from "@/notify";
import { generateRandomCode } from "@/utils/auth";

export default {
  name: "layout",
  components: {
    ContextMenu,
    Notifications,
    defaultBar,
    Sidebar,
    Prompts,
    Scrollbar,
    Tooltip,
  },
  data() {
    return {
      showContexts: true,
      dragCounter: 0,
      width: window.innerWidth,
      itemWeight: 0,
    };
  },
  mounted() {
    window.addEventListener("resize", this.updateIsMobile);
    if (state.user.themeColor) {
      document.documentElement.style.setProperty("--primaryColor", state.user.themeColor);
    }
    if (!state.sessionId) {
      mutations.setSession(generateRandomCode(8));
    }
    this.reEval()
    this.initialize();
  },
  computed: {
    isOnlyOffice() {
      return getters.currentView() === "onlyOfficeEditor";
    },
    scrollable() {
      return getters.isScrollable();
    },
    showPadding() {
      return getters.showBreadCrumbs() || getters.currentView() === "settings";
    },
    isLoggedIn() {
      return getters.isLoggedIn();
    },
    moveWithSidebar() {
      return getters.isSidebarVisible() && getters.isStickySidebar();
    },
    progress() {
      return getters.progress(); // Access getter directly from the store
    },
    currentPrompt() {
      return getters.currentPrompt(); // Access getter directly from the store
    },
    currentPromptName() {
      return getters.currentPromptName(); // Access getter directly from the store
    },
    req() {
      return state.req; // Access state directly from the store
    },
    user() {
      return state.user; // Access state directly from the store
    },
    showOverlay() {
      return getters.showOverlay();
    },
    isDarkMode() {
      return getters.isDarkMode();
    },
    currentView() {
      return getters.currentView();
    },
  },
  watch: {
    $route() {
      this.reEval()
    },
  },
  methods: {
    reEval() {
      mutations.setPreviewSource("");
      if (!getters.isLoggedIn()) {
        return;
      }
      const currentView = getters.currentView()
      mutations.setMultiple(false);
      if (getters.currentPromptName() !== "success") {
        mutations.closeHovers();
      }
      if (window.location.hash == "" && currentView == "listingView") {
        const element = document.getElementById("main");
        if (element) {
          element.scrollTop = 0;
        }
      }
    },
    async initialize() {
      if (getters.isLoggedIn()) {
        const sourceinfo = await filesApi.sources();
        mutations.updateSourceInfo(sourceinfo);
        if (state.user.permissions.realtime) {
          events.startSSE();
        }
        const maxUploads = state.user.fileLoading.maxConcurrentUpload;
        if (maxUploads > 10 || maxUploads < 1) {
          mutations.setMaxConcurrentUpload(1);
        }
      }
    },
    updateIsMobile() {
      mutations.setMobile();
    },
    resetPrompts() {
      mutations.closeSidebar();
      mutations.closeHovers();
      mutations.setSearch(false);
    },
  },
};
</script>

<style>
.scrollable {
  overflow: scroll !important;
  -webkit-overflow-scrolling: touch;
  /* Enable momentum scrolling in iOS */
}

.remove-padding-top {
  padding-top: 0 !important;
}

#main {
  overflow: unset;
  -ms-overflow-style: none;
  /* Internet Explorer 10+ */
  scrollbar-width: none;
  /* Firefox */
  transition: 0.5s ease;
}

#main.moveWithSidebar {
  padding-left: 20em;
}

#main::-webkit-scrollbar {
  display: none;
  /* Safari and Chrome */
}

#main > div {
  height: 100%;
}
</style>
