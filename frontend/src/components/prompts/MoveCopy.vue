<template>
    <div class="card floating">
        <div class="card-title">
            <h2>{{ $t(`prompts.${operation}`) }}</h2>
        </div>

        <div class="card-content">
            <p v-if="operation === 'copy'">{{ $t("prompts.copyMessage") }}</p>

            <file-list ref="fileList" @update:selected="updateDestination">
            </file-list>
        </div>

        <div class="card-action" style="display: flex; align-items: center; justify-content: space-between">
            <template v-if="user.permissions.modify">
                <button class="button button--flat" @click="$refs.fileList.createDir()"
                    :aria-label="$t('sidebar.newFolder')" :title="$t('sidebar.newFolder')" style="justify-self: left">
                    <span>{{ $t("sidebar.newFolder") }}</span>
                </button>
            </template>
            <div>
                <button class="button button--flat button--grey" @click="closeHovers" :aria-label="$t('buttons.cancel')"
                    :title="$t('buttons.cancel')">
                    {{ $t("buttons.cancel") }}
                </button>
                <button :disabled="destContainsSrc" class="button button--flat" @click="performOperation"
                    :aria-label="$t(`buttons.${operation}`)" :title="$t(`buttons.${operation}`)">
                    {{ $t(`buttons.${operation}`) }}
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import { mutations, state } from "@/store";
import FileList from "./FileList.vue";
import { filesApi } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { url } from "@/utils";
import { notify } from "@/notify";
import { removeLastDir } from "@/utils/url";

export default {
  name: "move-copy",
  components: { FileList },
  props: {
    operation: {
      type: String,
      required: true,
      validator: (value) => ["move", "copy"].includes(value),
    },
  },
  data: function () {
    return {
      current: window.location.pathname,
      destPath: "/", // Start at root of selected source
      destSource: null, // Will be set by FileList component
      items: [],
    };
  },
  computed: {
    destContainsSrc() {
      if (!this.destPath) {
        return false; // If dest is not set, we can't check containment
      }
      
      // Add null checks to prevent undefined errors
      if (!this.items || this.items.length === 0) {
        return false;
      }
      
      const itemPath = this.items[0]?.from;
      if (!itemPath) {
        return false; // If itemPath is undefined, we can't check containment
      }
      
      const parentDir = (this.operation === "copy" ? url.removeLastDir(itemPath) : removeLastDir(itemPath)) + "/";
      return this.destPath == parentDir || this.destPath.startsWith(itemPath);
    },
    user() {
      return state.user;
    },
    closeHovers() {
      return mutations.closeHovers();
    },
  },
  mounted() {
    if (state.isSearchActive) {
      // Add null checks to prevent undefined values
      if (state.selected && state.selected[0] && state.selected[0].path) {
        this.items = [
          {
            from: state.selected[0].path,
            fromSource: state.selected[0].source,
            name: state.selected[0].name,
          },
        ];
      }
    } else {
      if (state.selected && state.req && state.req.items) {
        for (let item of state.selected) {
          const reqItem = state.req.items[item];
          if (reqItem && reqItem.path) {
            this.items.push({
              from: reqItem.path,
              fromSource: state.req.source,
              name: reqItem.name,
            });
          }
        }
      }
    }
  },
  methods: {
    updateDestination(pathOrData) {
      // Handle both old format (just path) and new format (object with path and source)
      if (typeof pathOrData === 'string') {
        this.destPath = pathOrData;
        // For backward compatibility, keep the current source
        // This will be updated when FileList is modified to emit both
      } else if (pathOrData && pathOrData.path) {
        this.destPath = pathOrData.path;
        // Update destSource from FileList's selection
        this.destSource = pathOrData.source;
      }
    },
    performOperation: async function (event) {
      event.preventDefault();
      try {
        // Define the action function
        let action = async (overwrite, rename) => {
          for (let item of this.items) {
            // Ensure proper path construction without double slashes
            const destPath = this.destPath.endsWith('/') ? this.destPath : this.destPath + '/';
            item.to = destPath + item.name;
            // Always set toSource for cross-source operations
            item.toSource = this.destSource;
          }
          buttons.loading(this.operation);
          await filesApi.moveCopy(this.items, this.operation, overwrite, rename);
        };
        // Fetch destination files
        let dstResp = await filesApi.fetchFiles(
          this.destSource,
          this.destPath
        );
        let conflict = upload.checkConflict(this.items, dstResp.items);
        let overwrite = false;
        let rename = false;

        if (conflict) {
          await new Promise((resolve, reject) => {
            mutations.showHover({
              name: "replace-rename",
              confirm: async (event, option) => {
                overwrite = option == "overwrite";
                rename = option == "rename";
                event.preventDefault();
                try {
                  await action(overwrite, rename);
                  resolve(); // Resolve the promise if action succeeds
                } catch (e) {
                  reject(e); // Reject the promise if an error occurs
                }
              },
            });
          });
        } else {
          // Await the action call for non-conflicting cases
          await action(overwrite, rename);
        }
        mutations.closeHovers();
        mutations.setSearch(false);
        notify.showSuccess(`Server is ${this.operation === "copy" ? "copying" : "moving"} files...`);
      } catch (error) {
        notify.showError(error);
      }
    },
  },
};
</script>