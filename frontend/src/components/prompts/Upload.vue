<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.upload") }}</h2>
    </div>
    <div
      class="upload-prompt"
      @dragenter.prevent="onDragEnter"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      :class="{ 'drag-over': isDragging }"
    >
      <div class="upload-prompt-container">
        <i v-if="files.length === 0" class="material-icons">cloud_upload</i>
        <p v-if="files.length === 0">{{ $t("prompts.dragAndDrop") }}</p>
        <div class="button-group">
          <button @click="triggerFilePicker" class="button button--flat">
            {{ $t("buttons.file") }}
          </button>
          <button
            style="margin-left: 1em"
            @click="triggerFolderPicker"
            class="button button--flat"
          >
            {{ $t("buttons.folder") }}
          </button>
        </div>
      </div>
    </div>
    <div
      class="card-content"
      @drop.prevent="onDrop"
      :class="{ 'drag-over': isDragging }"
    >
      <div v-if="showConflictPrompt" class="conflict-overlay">
        <div class="card">
          <div class="card-content">
            <p>{{ $t("prompts.conflictsDetected") }}</p>
          </div>
          <div class="card-actions">
            <button
              @click="resolveConflict(false)"
              class="button button--flat button--grey"
            >
              {{ $t("buttons.cancel") }}
            </button>
            <button
              @click="resolveConflict(true)"
              class="button button--flat button--red"
            >
              {{ $t("buttons.replace") }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="files.length > 0" class="upload-list">
        <div v-for="file in files" :key="file.id" class="upload-item">
          <i class="material-icons file-icon">{{file.type === "directory" ? "folder" : "insert_drive_file" }}</i> <!-- eslint-disable-line @intlify/vue-i18n/no-raw-text -->
          <div class="file-info">
            <p class="file-name">{{ file.name }}</p>
            <progress-bar
              v-if="file.type !== 'directory'"
              :val="
                file.status === 'completed'
                  ? $t('prompts.completed')
                  : file.status === 'error'
                  ? $t('prompts.error')
                  : file.status === 'conflict'
                  ? $t('prompts.conflictsDetected')
                  : (file.progress / 100) * file.size
              "
              :unit="file.status === 'completed' || file.status === 'error' ? '' : 'bytes'"
              :max="file.size"
              :status="file.status"
              text-position="inside"
              size="20"
            >
            </progress-bar>
            <div v-else class="status-label">
              <span>{{ $t(`prompts.${file.status}`) }}</span>
            </div>
          </div>
          <div class="file-actions">
            <button
              v-if="file.status === 'uploading'"
              @click="uploadManager.pause(file.id)"
              class="action"
              :aria-label="$t('buttons.pause')"
              :title="$t('buttons.pause')"
            >
              <i class="material-icons">pause</i>
            </button>
            <button
              v-if="file.status === 'paused'"
              @click="uploadManager.resume(file.id)"
              class="action"
              :aria-label="$t('buttons.resume')"
              :title="$t('buttons.resume')"
            >
              <i class="material-icons">play_arrow</i>
            </button>
            <button
              v-if="file.status === 'error'"
              @click="uploadManager.retry(file.id)"
              class="action"
              :aria-label="$t('buttons.retry')"
              :title="$t('buttons.retry')"
            >
              <i class="material-icons">replay</i>
            </button>
            <button
              v-if="file.status === 'conflict'"
              @click="handleConflictAction(file)"
              class="action"
              :aria-label="$t('buttons.replace')"
              :title="$t('buttons.replace')"
            >
              <i class="material-icons">sync_problem</i>
            </button>
            <button
              @click="cancelUpload(file.id)"
              class="action"
              :aria-label="$t('buttons.cancel')"
              :title="$t('buttons.cancel')"
            >
              <i class="material-icons">close</i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="card-actions">
      <button
        @click="close"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.close") }}
      </button>
      <div class="spacer"></div>
      <button
        v-if="canPauseAll"
        @click="uploadManager.pauseAll"
        class="button button--flat"
        :aria-label="$t('buttons.pauseAll')"
        :title="$t('buttons.pauseAll')"
      >
        {{ $t("buttons.pauseAll") }}
      </button>
      <button
        v-if="canResumeAll"
        @click="uploadManager.resumeAll"
        class="button button--flat"
        :aria-label="$t('buttons.resumeAll')"
        :title="$t('buttons.resumeAll')"
      >
        {{ $t("buttons.resumeAll") }}
      </button>
      <button
        @click="clearCompleted"
        class="button button--flat"
        :disabled="!hasCompleted"
        :aria-label="$t('buttons.clearCompleted')"
        :title="$t('buttons.clearCompleted')"
      >
        {{ $t("buttons.clearCompleted") }}
      </button>
    </div>

    <input
      ref="fileInput"
      @change="onFilePicked"
      type="file"
      multiple
      style="display: none"
    />
    <input
      ref="folderInput"
      @change="onFolderPicked"
      type="file"
      webkitdirectory
      directory
      multiple
      style="display: none"
    />
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { uploadManager } from "@/utils/upload";
import { mutations, state } from "@/store";
import ProgressBar from "@/components/ProgressBar.vue";

export default {
  name: "UploadFiles",
  components: {
    ProgressBar,
  },
  props: {
    initialItems: {
      type: Object,
      default: null,
    },
    filesToReplace: {
      type: Array,
      default: () => [],
    },
  },
  setup(props) {
    console.log("Upload.vue: setup started. Props:", props);
    const fileInput = ref(null);
    const folderInput = ref(null);
    const files = computed(() => uploadManager.queue);
    const isDragging = ref(false);
    const showConflictPrompt = ref(false);
    let conflictResolver = null;

    let wakeLock = null;

    const handleConflict = (resolver) => {
      showConflictPrompt.value = true;
      conflictResolver = resolver;
    };

    const resolveConflict = (overwrite) => {
      if (conflictResolver) {
        conflictResolver(overwrite);
      }
      showConflictPrompt.value = false;
      conflictResolver = null;
    };

    const acquireWakeLock = async () => {
      if (!("wakeLock" in navigator)) {
        return;
      }
      try {
        if (wakeLock !== null) return; // Already locked
        wakeLock = await navigator.wakeLock.request("screen");
        wakeLock.addEventListener("release", () => {
          wakeLock = null;
        });
      } catch (err) {
        console.error(`Wake Lock failed: ${err.name}, ${err.message}`);
      }
    };

    const releaseWakeLock = () => {
      if (wakeLock !== null) {
        wakeLock.release();
        wakeLock = null;
      }
    };

    const isUploading = computed(() =>
      files.value.some(
        (file) => file.status === "pending" || file.status === "uploading"
      )
    );

    watch(isUploading, (active) => {
      if (active) {
        acquireWakeLock();
      } else {
        releaseWakeLock();
      }
    });

    const hasCompleted = computed(() =>
      files.value.some((file) => file.status === "completed")
    );

    const canPauseAll = computed(() =>
      files.value.some((file) => file.status === "uploading")
    );

    const canResumeAll = computed(
      () =>
        !canPauseAll.value &&
        files.value.some((file) => file.status === "paused")
    );

    const close = () => {
      mutations.closeHovers();
    };

    const clearCompleted = () => {
      uploadManager.clearCompleted();
    };

    const handleVisibilityChange = async () => {
      if (document.visibilityState === "visible" && isUploading.value) {
        acquireWakeLock();
      }
    };

    onMounted(async () => {
      document.addEventListener("visibilitychange", handleVisibilityChange);
      uploadManager.setOnConflict(handleConflict);
      if (props.initialItems) {
        await processItems(props.initialItems);
      }
    });

    onUnmounted(() => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      uploadManager.setOnConflict(() => {}); // cleanup
      releaseWakeLock();
    });

    const triggerFilePicker = () => {
      if (fileInput.value) fileInput.value.click();
    };

    const triggerFolderPicker = () => {
      if (folderInput.value) folderInput.value.click();
    };

    const onFilePicked = (event) => {
      const pickedFiles = event.target.files;
      if (pickedFiles.length > 0) {
        processFileList(pickedFiles);
      }
      if (event.target) event.target.value = null;
    };

    const onFolderPicked = (event) => {
      const pickedFiles = event.target.files;
      if (pickedFiles.length > 0) {
        processFileList(pickedFiles);
      }
      if (event.target) event.target.value = null;
    };

    const onDrop = async (event) => {
      isDragging.value = false;
      if (event.dataTransfer.items) {
        const items = Array.from(event.dataTransfer.items);
        console.log("Upload.vue: Processing dropped items:", items);
        await processDroppedItems(items);
      } else {
        const droppedFiles = event.dataTransfer.files;
        console.log("Upload.vue: Processing dropped files (fallback):", droppedFiles);
        processFileList(droppedFiles);
      }
    };

    const onDragEnter = () => {
      isDragging.value = true;
    };

    const onDragOver = () => {
      isDragging.value = true;
    };

    const onDragLeave = () => {
      isDragging.value = false;
    };

    const getFilesFromDirectoryEntry = async (entry) => {
      if (entry.isFile) {
        return new Promise((resolve) => {
          entry.file((file) => {
            const relativePath = entry.fullPath.startsWith("/")
              ? entry.fullPath.substring(1)
              : entry.fullPath;
            resolve([{ file, relativePath }]);
          });
        });
      }
      if (entry.isDirectory) {
        const reader = entry.createReader();
        const entries = await new Promise((resolve) => {
          reader.readEntries((e) => resolve(e));
        });
        const allFiles = await Promise.all(
          entries.map((subEntry) => getFilesFromDirectoryEntry(subEntry))
        );
        return allFiles.flat();
      }
      return [];
    };

    const processDroppedItems = async (items) => {
      const filesToUpload = [];
      for (const item of items) {
        const entry = item.webkitGetAsEntry();
        if (entry) {
          const files = await getFilesFromDirectoryEntry(entry);
          filesToUpload.push(...files);
        }
      }

      if (filesToUpload.length > 0) {
        uploadManager.add(state.req.path, filesToUpload);
      }
    };

    const processFileList = (fileList) => {
      const filesToAdd = Array.from(fileList).map((file) => ({
        file,
        relativePath: file.webkitRelativePath || file.name,
      }));
      if (filesToAdd.length > 0) {
        uploadManager.add(state.req.path, filesToAdd);
      }
    };

    const handleConflictAction = (file) => {
      mutations.showHover({
        name: "replace",
        confirm: () => {
          uploadManager.retry(file.id, true);
          mutations.closeTopHover();
        },
      });
    };

    const cancelUpload = (id) => {
      uploadManager.cancel(id);
    };

    return {
      triggerFilePicker,
      triggerFolderPicker,
      onFilePicked,
      onFolderPicked,
      fileInput,
      folderInput,
      files,
      isDragging,
      onDragEnter,
      onDragLeave,
      onDragOver,
      onDrop,
      cancelUpload,
      uploadManager,
      close,
      clearCompleted,
      hasCompleted,
      showConflictPrompt,
      resolveConflict,
      canPauseAll,
      canResumeAll,
      handleConflictAction,
    };
  },
};
</script>

<style scoped>
.upload-prompt {
  text-align: center;
  padding: 2em;
  border: 2px dashed #ccc;
  border-radius: 8px;
  margin: 0 1em;
}

.upload-prompt.drag-over {
  background-color: #f0f0f0;
  border-color: #333;
}

.upload-prompt-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-prompt i {
  font-size: 4em;
  color: #ccc;
}

.upload-list {
  overflow-y: auto;
  padding-right: 0.5em; /* To avoid scrollbar overlapping content */
  flex-grow: 1;
  display: flex;
  flex-direction: column-reverse;
  min-height: 0;
}

.upload-item {
  display: flex;
}

.file-icon {
  margin-right: 0.5em;
  color: #999;
}

.file-info {
  flex-grow: 1;
}

.file-name {
  margin: 0;
  font-size: 0.9em;
}

.status-label {
  color: #777;
  font-size: 0.8em;
  margin-top: 5px;
}

.file-actions .action {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.2em;
}

.file-actions .action i {
  font-size: 1.2em;
}

.drag-over {
  border-color: #2196f3;
  background: #f4f4f4;
}

.card-actions {
  display: flex;
  justify-content: space-between;
  padding: 0.5em;
}

.spacer {
  flex-grow: 1;
}

.card.floating {
  display: flex !important;
  flex-direction: column;
  max-height: 85vh;
}

.card-content {
  flex-grow: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
  padding: 1em;
}

.conflict-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  z-index: 999;
  display: flex;
  justify-content: center;
  align-items: center;
}

.conflict-overlay .card {
  background-color: var(--card-background-color);
  padding: 1em;
  border-radius: 8px;
}
</style>
