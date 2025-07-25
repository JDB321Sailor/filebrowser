<template>
  <div>
    <component
      v-if="showOverlay"
      :ref="currentPromptName"
      :is="currentPromptName"
      v-bind="currentPromptProps"
    />
  </div>
</template>

<script>
import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import Rename from "./Rename.vue";
import Download from "./Download.vue";
import MoveCopy from "./MoveCopy.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import ReplaceRename from "./ReplaceRename.vue";
import Share from "./Share.vue";
import Upload from "./Upload.vue";
import ShareDelete from "./ShareDelete.vue";
import DeleteUser from "./DeleteUser.vue";
import CreateApi from "./CreateApi.vue";
import ActionApi from "./ActionApi.vue";
import Sidebar from "../sidebar/Sidebar.vue";
import buttons from "@/utils/buttons";
import Totp from "./Totp.vue";
import { state, getters, mutations } from "@/store"; // Import your custom store

export default {
  name: "prompts",
  components: {
    Info,
    Delete,
    Rename,
    Download,
    Move: MoveCopy,
    Copy: MoveCopy,
    Share,
    NewFile,
    NewDir,
    Help,
    Replace,
    ReplaceRename,
    Totp,
    Upload,
    ShareDelete,
    Sidebar,
    DeleteUser,
    CreateApi,
    ActionApi,
  },
  data() {
    return {
      pluginData: {
        buttons,
        store: state, // Directly use state
        router: this.$router,
      },
    };
  },
  created() {
    window.addEventListener("keydown", (event) => {
      let currentPrompt = getters.currentPrompt();
      if (!currentPrompt) return;

      let prompt = this.$refs[currentPrompt.name];

      // Esc!
      if (event.keyCode === 27) {
        event.stopImmediatePropagation();
        mutations.closeHovers();
      } else if (event.KeyCode === 8) {
        event.preventDefault();
      }
      // Enter
      if (event.keyCode === 13) {
        switch (currentPrompt.name) {
          case "delete":
            prompt.submit();
            break;
          case "copy":
          case "move":
            prompt.performOperation(event);
            break;
          case "replace":
            prompt.showConfirm(event);
            break;
        }
      }
    });
  },
  computed: {
    currentPromptName() {
      if (getters.currentPromptName() == null) {
        return "";
      }
      return getters.currentPromptName();
    },
    currentPrompt() {
      if (getters.currentPrompt() == null) {
        return {
          props: {},
        };
      }
      return getters.currentPrompt();
    },
    currentPromptProps() {
      const prompt = this.currentPrompt;
      const promptName = this.currentPromptName;
      // For move and copy prompts, add the operation prop
      if (promptName === "move" || promptName === "copy") {
        return {
          ...prompt.props,
          operation: promptName,
        };
      }
      return prompt.props;
    },
    plugins() {
      return state.plugins;
    },
    showOverlay() {
      return getters.currentPromptName() !== "more";
    },
  },
  methods: {},
};
</script>
