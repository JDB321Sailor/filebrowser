<template>
  <div class="no-select">
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
    <div v-else>
      <div
        id="listingView"
        ref="listingView"
        class="font-size-large"
        :class="{ 'add-padding': isStickySidebar, [listingViewMode]: true }"
        v-if="numDirs + numFiles == 0"
      >
        <h2 class="message">
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{ $t("files.lonely") }}</span>
        </h2>
        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />
      </div>
      <div
        v-else
        id="listingView"
        ref="listingView"
        :class="{
          'add-padding': isStickySidebar,
          [listingViewMode]: true,
          dropping: isDragging,
        }"
        class="file-icons"
      >
        <div>
          <div class="header" :class="{ 'dark-mode-item-header': isDarkMode }">
            <p
              :class="{ active: nameSorted }"
              class="name"
              role="button"
              tabindex="0"
              @click="sort('name')"
              :title="$t('files.sortByName')"
              :aria-label="$t('files.sortByName')"
            >
              <span>{{ $t("files.name") }}</span>
              <i class="material-icons">{{ nameIcon }}</i>
            </p>

            <p
              :class="{ active: sizeSorted }"
              class="size"
              role="button"
              tabindex="0"
              @click="sort('size')"
              :title="$t('files.sortBySize')"
              :aria-label="$t('files.sortBySize')"
            >
              <span>{{ $t("files.size") }}</span>
              <i class="material-icons">{{ sizeIcon }}</i>
            </p>
            <p
              :class="{ active: modifiedSorted }"
              class="modified"
              role="button"
              tabindex="0"
              @click="sort('modified')"
              :title="$t('files.sortByLastModified')"
              :aria-label="$t('files.sortByLastModified')"
            >
              <span>{{ $t("files.lastModified") }}</span>
              <i class="material-icons">{{ modifiedIcon }}</i>
            </p>
          </div>
        </div>
        <div v-if="numDirs > 0">
          <div class="header-items">
            <h2>{{ $t("files.folders") }}</h2>
          </div>
        </div>
        <div
          v-if="numDirs > 0"
          class="folder-items"
          :class="{ lastGroup: numFiles === 0 }"
        >
          <item
            v-for="item in dirs"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.type == 'directory'"
            v-bind:source="req.source"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:reducedOpacity="item.hidden || isDragging"
          />
        </div>
        <div v-if="numFiles > 0">
          <div class="header-items">
            <h2>{{ $t("files.files") }}</h2>
          </div>
        </div>
        <div v-if="numFiles > 0" class="file-items" :class="{ lastGroup: numFiles > 0 }">
          <item
            v-for="item in files"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isDir="item.type == 'directory'"
            v-bind:modified="item.modified"
            v-bind:source="req.source"
            v-bind:type="item.type"
            v-bind:size="item.size"
            v-bind:path="item.path"
            v-bind:reducedOpacity="item.hidden || isDragging"
          />
        </div>

        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />
      </div>
    </div>
  </div>
</template>

<script>
import downloadFiles from "@/utils/download";
import { filesApi } from "@/api";
import { router } from "@/router";
import * as upload from "@/utils/upload";
import throttle from "@/utils/throttle";
import { state, mutations, getters } from "@/store";
import { url } from "@/utils";

import Item from "@/components/files/ListingItem.vue";
export default {
  name: "listingView",
  components: {
    Item,
  },
  data() {
    return {
      columnWidth: 250 + state.user.gallerySize * 50,
      dragCounter: 0,
      width: window.innerWidth,
      lastSelected: {}, // Add this to track the currently focused item
      contextTimeout: null, // added for safari context menu
      ctrKeyPressed: false,
    };
  },
  watch: {
    gallerySize() {
      this.columnWidth = 250 + state.user.gallerySize * 50;
      this.colunmsResize();
    },
    scrolling() {
      const scrollContainer = this.$refs.listingView;
      if (!scrollContainer) return;

      // Select all visible listing items
      const itemNodes = scrollContainer.querySelectorAll(".listing-item");

      // Find the first item near the top of the viewport
      let topItem = null;
      let minTop = Infinity;
      itemNodes.forEach((el) => {
        const rect = el.getBoundingClientRect();
        if (rect.top >= 0 && rect.top < minTop) {
          minTop = rect.top;
          topItem = el;
        }
      });

      if (!topItem) return;

      const letter = topItem.getAttribute("data-name")?.[0]?.toUpperCase() || "A";
      let category = "folders"; // Default category
      if (this.numFiles > 0) {
        // Decide category by checking which section is above
        const fileSection = this.$el.querySelector(".file-items");
        const fileTop = fileSection?.getBoundingClientRect().top ?? 0;
        category = fileTop <= 0 ? "files" : "folders";
      }
      if (this.numDirs == 0) {
        category = "files"; // If no directories, only files
      }

      mutations.updateListing({
        ...state.listing,
        category,
        letter,
      });
    },
  },
  computed: {
    isDragging() {
      return this.dragCounter > 0;
    },
    scrolling() {
      return state.listing.scrollRatio;
    },
    isStickySidebar() {
      return getters.isStickySidebar();
    },
    lastFolderIndex() {
      const allItems = [...this.items.dirs, ...this.items.files];
      for (let i = 0; i < allItems.length; i++) {
        if (allItems[i].type != "directory") {
          return i - 1;
        }
      }
      if (allItems.length > 0) {
        return allItems.length;
      }

      return null; // Return null if there are no files
    },
    numColumns() {
      if (!getters.isCardView()) {
        return 1;
      }
      const elem = document.querySelector("#main");
      if (!elem) {
        return 1;
      }
      let columns = Math.floor(elem.offsetWidth / this.columnWidth);
      if (columns === 0) columns = 1;
      return columns;
    },
    // Create a computed property that references the Vuex state
    gallerySize() {
      return state.user.gallerySize;
    },
    isDarkMode() {
      return state.user?.darkMode;
    },
    getMultiple() {
      return state.multiple;
    },
    nameSorted() {
      return state.user.sorting.by === "name";
    },
    sizeSorted() {
      return state.user.sorting.by === "size";
    },
    modifiedSorted() {
      return state.user.sorting.by === "modified";
    },
    ascOrdered() {
      return state.user.sorting.asc;
    },
    items() {
      return getters.reqItems();
    },
    numDirs() {
      return getters.reqNumDirs();
    },
    numFiles() {
      return getters.reqNumFiles();
    },
    dirs() {
      return this.items.dirs;
    },
    files() {
      return this.items.files;
    },
    nameIcon() {
      if (this.nameSorted && !this.ascOrdered) {
        return "arrow_upward";
      }

      return "arrow_downward";
    },
    sizeIcon() {
      if (this.sizeSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    modifiedIcon() {
      if (this.modifiedSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    viewIcon() {
      const icons = {
        list: "view_module",
        compact: "view_module",
        normal: "grid_view",
        gallery: "view_list",
      };
      return icons[state.user.viewMode];
    },
    listingViewMode() {
      this.colunmsResize();
      return state.user.viewMode;
    },
    selectedCount() {
      return state.selected.length;
    },
    req() {
      return state.req;
    },
    loading() {
      return getters.isLoading();
    },
  },
  mounted() {
    mutations.setSearch(false);
    this.lastSelected = state.selected;
    // Check the columns size for the first time.
    this.colunmsResize();
    // Add the needed event listeners to the window and document.
    window.addEventListener("keydown", this.keyEvent);
    window.addEventListener("resize", this.windowsResize);
    window.addEventListener("click", this.clickClear);
    window.addEventListener("keyup", this.clearCtrKey);
    window.addEventListener("dragover", this.preventDefault);
    this.$el.addEventListener("touchmove", this.handleTouchMove);

    this.$el.addEventListener("contextmenu", this.openContext);
    // Adjust contextmenu listener based on browser
    if (state.isSafari) {
      // For Safari, add touchstart or mousedown to open the context menu
      this.$el.addEventListener("touchstart", this.openContextForSafari, {
        passive: true,
      });
      this.$el.addEventListener("mousedown", this.openContextForSafari);

      // Also clear the timeout if the user clicks or taps quickly
      this.$el.addEventListener("touchend", this.cancelContext);
      this.$el.addEventListener("mouseup", this.cancelContext);
    }

    // if safari , make sure click and hold opens context menu, but not for any other browser
    if (!state.user.permissions?.modify) return;
    this.$el.addEventListener("dragenter", this.dragEnter);
    this.$el.addEventListener("dragleave", this.dragLeave);
    this.$el.addEventListener("drop", this.drop);
  },
  beforeUnmount() {
    // Remove event listeners before destroying this page.
    window.removeEventListener("keydown", this.keyEvent);
    window.removeEventListener("resize", this.windowsResize);
    window.removeEventListener("click", this.clickClear);
    window.removeEventListener("keyup", this.clearCtrKey);
    window.removeEventListener("dragover", this.preventDefault);

    this.$el.removeEventListener("touchmove", this.handleTouchMove);
    this.$el.removeEventListener("contextmenu", this.openContext);

    // If Safari, remove touch/mouse listeners
    if (state.isSafari) {
      this.$el.removeEventListener("touchstart", this.openContextForSafari);
      this.$el.removeEventListener("mousedown", this.openContextForSafari);
      this.$el.removeEventListener("touchend", this.cancelContext);
      this.$el.removeEventListener("mouseup", this.cancelContext);
    }

    // Also clean up drag/drop listeners on the component's root element
    if (state.user.permissions?.modify) {
      this.$el.removeEventListener("dragenter", this.dragEnter);
      this.$el.removeEventListener("dragleave", this.dragLeave);
      this.$el.removeEventListener("drop", this.drop);
    }
  },
  methods: {
    cancelContext() {
      if (this.contextTimeout) {
        clearTimeout(this.contextTimeout);
        this.contextTimeout = null;
      }
      this.isLongPress = false;
    },
    openContextForSafari(event) {
      this.cancelContext(); // Clear any previous timeouts
      this.isLongPress = false; // Reset state
      this.isSwipe = false; // Reset swipe detection

      const touch = event.touches[0];
      this.touchStartX = touch.clientX;
      this.touchStartY = touch.clientY;

      // Start the long press detection
      this.contextTimeout = setTimeout(() => {
        if (!this.isSwipe) {
          this.isLongPress = true;
          event.preventDefault(); // Suppress Safari's callout menu
          this.openContext(event); // Open the custom context menu
        }
      }, 500); // Long press delay (adjust as needed)
    },
    handleTouchMove(event) {
      const touch = event.touches[0];
      const deltaX = Math.abs(touch.clientX - this.touchStartX);
      const deltaY = Math.abs(touch.clientY - this.touchStartY);
      // Set a threshold for movement to detect a swipe
      const movementThreshold = 10; // Adjust as needed
      if (deltaX > movementThreshold || deltaY > movementThreshold) {
        this.isSwipe = true;
        this.cancelContext(); // Cancel long press if swipe is detected
      }
    },
    handleTouchEnd() {
      this.cancelContext(); // Clear timeout
      this.isSwipe = false; // Reset swipe state
    },
    base64(name) {
      return url.base64Encode(name);
    },
    // Helper method to select the first item if nothing is selected
    selectFirstItem() {
      mutations.resetSelected();
      const allItems = [...this.items.dirs, ...this.items.files];
      if (allItems.length > 0) {
        mutations.addSelected(allItems[0].index);
      }
    },
    // Helper method to select an item by index
    selectItem(index) {
      mutations.resetSelected();
      mutations.addSelected(index);
    },
    // Helper method to handle selection based on arrow keys
    navigateKeboardArrows(arrowKey) {
      let selectedIndex = state.selected.length > 0 ? state.selected[0] : null;

      if (selectedIndex === null) {
        // If nothing is selected, select the first item
        this.selectFirstItem();
        return;
      }

      const allItems = [...this.items.dirs, ...this.items.files]; // Combine files and directories

      // Find the current index of the selected item
      let currentIndex = allItems.findIndex((item) => item.index === selectedIndex);

      // If no item is selected, select the first item
      if (currentIndex === -1) {
        // Check if there are any items to select
        if (allItems.length > 0) {
          currentIndex = 0;
          this.selectItem(allItems[currentIndex].index);
        }
        return;
      }
      let newSelected = null;
      const fileSelected = currentIndex > this.lastFolderIndex;
      const nextIsDir = currentIndex - this.numColumns <= this.lastFolderIndex;
      const folderSelected = currentIndex <= this.lastFolderIndex;
      const nextIsFile = currentIndex + this.numColumns > this.lastFolderIndex;
      const nextHopExists = currentIndex + this.numColumns < allItems.length;
      const thisColumnNum =
        ((currentIndex - this.lastFolderIndex - 1) % this.numColumns) + 1;
      const lastFolderColumn = (this.lastFolderIndex % this.numColumns) + 1;
      const thisColumnNum2 = (currentIndex + 1) % this.numColumns;
      let firstRowColumnPos = this.lastFolderIndex + thisColumnNum2;
      let newPos = currentIndex - lastFolderColumn;
      switch (arrowKey) {
        case "ArrowUp":
          if (currentIndex - this.numColumns < 0) {
            // do nothing
            break;
          }
          if (!getters.isCardView) {
            newSelected = allItems[currentIndex - 1].index;
            break;
          }
          // do normal move
          if (!(fileSelected && nextIsDir)) {
            newSelected = allItems[currentIndex - this.numColumns].index;
            break;
          }

          // complex logic to move from files to folders
          if (lastFolderColumn < thisColumnNum) {
            newPos -= this.numColumns;
          }
          newSelected = allItems[newPos].index;

          break;

        case "ArrowDown":
          if (currentIndex >= allItems.length) {
            // do nothing - last item
            break;
          }
          if (!getters.isCardView) {
            newSelected = allItems[currentIndex + 1].index;
            break;
          }
          if (!nextHopExists) {
            // do nothing - next item is out of bounds
            break;
          }

          if (!(folderSelected && nextIsFile)) {
            newSelected = allItems[currentIndex + this.numColumns].index;
            break;
          }
          // complex logic for moving from folders to files
          if (firstRowColumnPos <= this.lastFolderIndex) {
            firstRowColumnPos += this.numColumns;
          }
          newSelected = allItems[firstRowColumnPos].index;
          break;

        case "ArrowLeft":
          if (currentIndex > 0) {
            newSelected = allItems[currentIndex - 1].index;
          }
          break;

        case "ArrowRight":
          if (currentIndex < allItems.length - 1) {
            newSelected = allItems[currentIndex + 1].index;
          }
          break;
      }
      if (newSelected != null) {
        this.selectItem(newSelected);
        setTimeout(() => {
          // Find the element with class "item" and aria-selected="true"
          const element = document.querySelector('.item[aria-selected="true"]');
          // Scroll the element into view if it exists
          if (element) {
            element.scrollIntoView({
              behavior: "smooth",
              block: "end",
              inline: "nearest",
            });
          }
        }, 50);
      }
    },
    clearCtrKey(event) {
      const { ctrlKey, metaKey } = event;
      const modifierKeys = ctrlKey || metaKey;
      if (!modifierKeys) {
        this.ctrKeyPressed = false;
      }
    },
    keyEvent(event) {
      if (state.isSearchActive || getters.currentView() != "listingView" || getters.currentPromptName() != null) {
        return;
      }
      const { key, ctrlKey, metaKey, which } = event;
      // Check if the key is alphanumeric
      const isAlphanumeric = /^[a-z0-9]$/i.test(key);
      const modifierKeys = ctrlKey || metaKey;
      if (isAlphanumeric && !modifierKeys && getters.currentPromptName() == null) {
        this.alphanumericKeyPress(key); // Call the alphanumeric key press function
        return;
      }
      if (!modifierKeys && getters.currentPromptName() != null) {
        return;
      }
      // Handle the space bar key
      if (key === " ") {
        event.preventDefault();
        if (state.isSearchActive) {
          mutations.setSearch(false);
          mutations.closeHovers();
        } else {
          mutations.setSearch(true);
        }
      }
      if (getters.currentPromptName() != null) {
        return;
      }
      let currentPath = state.route.path.replace(/\/+$/, ""); // Remove trailing slashes
      let newPath = currentPath.substring(0, currentPath.lastIndexOf("/"));

      if (modifierKeys) {
        this.ctrKeyPressed = true;
        return;
      }

      // Handle key events using a switch statement
      switch (key) {
        case "Enter":
          if (this.selectedCount === 1) {
            const selected = getters.getFirstSelected();
            const selectedUrl = url.buildItemUrl(selected.source, selected.path)
            router.push({ path: selectedUrl });
          }
          break;

        case "Backspace":
          if (getters.CurrentPromptName !== null) {
            return;
          }
          // go back
          router.push({ path: newPath });
          break;

        case "Escape":
          mutations.resetSelected();
          break;

        case "Delete":
          if (!state.user.permissions.modify || state.selected.length === 0) return;
          mutations.showHover("delete");
          break;

        case "F2":
          if (!state.user.permissions.modify || state.selected.length !== 1) return;
          mutations.showHover("rename");
          break;

        case "ArrowUp":
        case "ArrowDown":
        case "ArrowLeft":
        case "ArrowRight":
          event.preventDefault();
          this.navigateKeboardArrows(key);
          break;
      }

      const charKey = String.fromCharCode(which).toLowerCase();

      switch (charKey) {
        case "c":
        case "x":
          this.copyCut(event, charKey);
          break;
        case "v":
          this.paste(event);
          break;
        case "a":
          event.preventDefault();
          this.selectAll();
          break;
        case "s":
          event.preventDefault();
          downloadFiles(state.selected);
          break;
      }
    },
    selectAll() {
      for (let file of this.items.files) {
        if (state.selected.indexOf(file.index) === -1) {
          mutations.addSelected(file.index);
        }
      }
      for (let dir of this.items.dirs) {
        if (state.selected.indexOf(dir.index) === -1) {
          mutations.addSelected(dir.index);
        }
      }
    },
    alphanumericKeyPress(key) {
      // Convert the key to uppercase to match the case-insensitive search
      const searchLetter = key.toLowerCase();
      const currentSelected = getters.getFirstSelected();
      let currentName = null;
      let findNextWithName = false;

      if (currentSelected != undefined) {
        currentName = currentSelected.name.toLowerCase();
        if (currentName.startsWith(searchLetter)) {
          findNextWithName = true;
        }
      }
      // Combine directories and files (assuming they are stored in this.items.dirs and this.items.files)
      const allItems = [...this.items.dirs, ...this.items.files];
      let foundPrevious = false;
      let firstFound = null;
      // Iterate over all items to find the first one where the name starts with the searchLetter
      for (let i = 0; i < allItems.length; i++) {
        const itemName = allItems[i].name.toLowerCase();
        if (!itemName.startsWith(searchLetter)) {
          continue;
        }
        if (firstFound == null) {
          firstFound = allItems[i].index;
        }
        if (!findNextWithName) {
          // return first you find
          this.selectItem(allItems[i].index);
          return;
        }
        if (itemName == currentName) {
          foundPrevious = true;
          continue;
        }
        if (foundPrevious) {
          this.selectItem(allItems[i].index);
          return;
        }
      }
      // select the first item again
      if (firstFound != null) {
        this.selectItem(firstFound);
      }
    },
    preventDefault(event) {
      // Wrapper around prevent default.
      event.preventDefault();
    },
    copyCut(event, key) {
      if (event.target.tagName.toLowerCase() === "input") {
        return;
      }

      let items = state.selected.map((i) => ({
        from: state.req.items[i].path,
        fromSource: state.req.items[i].source,
        name: state.req.items[i].name,
      }));

      if (items.length === 0) {
        return;
      }

      this.clipboard = {
        key: key,
        items: items,
        path: state.route.path,
      };
    },
    async paste(event) {
      if (event.target.tagName.toLowerCase() === "input") {
        return;
      }

      let items = this.clipboard.items.map((item) => ({
        from: item.from,
        fromSource: item.fromSource,
        to: state.route.path + item.name,
        toSource: item.toSource
      }));

      if (items.length === 0) {
        return;
      }
      mutations.setLoading("listing", true);
      let action = async (overwrite, rename) => {
        await filesApi.moveCopy(items, "copy", overwrite, rename);
        mutations.setLoading("listing", false);
      };

      if (this.clipboard.key === "x") {
        action = async (overwrite, rename) => {
          await filesApi.moveCopy(items, "move", overwrite, rename);

          this.clipboard = {};
          mutations.setLoading("listing", false);
        };
      }

      if (this.clipboard.path === state.route.path) {
        action(false, true);
        return;
      }

      const conflict = upload.checkConflict(items, state.req.items);

      if (conflict) {
        this.currentPrompt = {
          name: "replace-rename",
          confirm: (event, option) => {
            const overwrite = option === "overwrite";
            const rename = option === "rename";

            event.preventDefault();
            mutations.closeHovers();
            action(overwrite, rename);
          },
        };
        return;
      }

      action(false, false);
    },
    colunmsResize() {
      document.documentElement.style.setProperty(
        "--item-width",
        `calc(${100 / this.numColumns}% - 1em)`
      );

      if (state.user.viewMode == "gallery") {
        document.documentElement.style.setProperty(
          "--item-height",
          `calc(${this.columnWidth / 25}em)`
        );
      } else {
        document.documentElement.style.setProperty("--item-height", `auto`);
      }
    },
    dragEnter(event) {
      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );
      if (isInternal) {
        return;
      }
      this.dragCounter++;
    },
    dragLeave(event) {
      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );
      if (isInternal) {
        return;
      }
      if (this.dragCounter == 0) {
        return;
      }
      this.dragCounter--;
    },
    async drop(event) {
      const isInternal = Array.from(event.dataTransfer.types).includes(
        "application/x-filebrowser-internal-drag"
      );

      if (isInternal) {
        return;
      }
      this.handleDrop(event);
    },
    async uploadInput(event) {
      this.handleDrop(event);
    },
    sort(field) {
      let asc = false;
      if (
        (field === "name" && this.nameIcon === "arrow_upward") ||
        (field === "size" && this.sizeIcon === "arrow_upward") ||
        (field === "modified" && this.modifiedIcon === "arrow_upward")
      ) {
        asc = true;
      }
      // Commit the updateSort mutation
      mutations.updateListingSortConfig({ field, asc });
      mutations.updateListingItems();
      this.lastSelected = state.selected;
    },
    setMultiple(val) {
      mutations.setMultiple(val == true);
      showMultipleSelection();
    },
    openSearch() {
      this.currentPrompt = "search";
    },
    windowsResize: throttle(function () {
      this.colunmsResize();
      this.width = window.innerWidth;
      // Listing element is not displayed
      if (this.$refs.listingView == null) return;
    }, 100),
    openContext(event) {
      event.preventDefault();
      event.stopPropagation();
      mutations.showHover({
        name: "ContextMenu",
        props: {
          showCentered: getters.isMobile(),
          posX: event.clientX,
          posY: event.clientY,
        },
      });
    },
    clickClear(event) {
      // if control or shift is pressed, do not clear the selection
      if (this.ctrKeyPressed || event.shiftKey) return;
      const sameAsBefore = state.selected == this.lastSelected;
      if (sameAsBefore && !state.multiple && getters.currentPromptName() == null) {
        mutations.resetSelected();
      }
      this.lastSelected = state.selected;
    },
    async handleDrop(event) {
      event.preventDefault();
      this.dragCounter = 0;

      if (event.type === "drop") {
        mutations.showHover({
          name: "upload",
          props: {
            initialItems: event.dataTransfer.items,
          },
        });
      } else {
        // This is for the <input type="file"> fallback
        const files = event.target.files;
        if (!files || files.length === 0) {
          return;
        }

        mutations.showHover({
          name: "upload",
          props: {
            // we send it as an array-like object so that it can be processed like a FileList by the Upload component
            initialItems: Array.from(files),
          },
        });
      }
    },
  },
};
</script>

<style>
.dark-mode-item-header {
  border-color: var(--divider) !important;
  background: var(--surfacePrimary) !important;
  user-select: none;
}

.header-items {
  width: 100% !important;
  max-width: 100% !important;
  justify-content: center;
}

.add-padding {
  padding-left: 0.5em;
}
.font-size-large h2 {
  font-size: 2em !important;
}

#listingView.dropping {
  transform: scale(0.97);
  border-radius: 1em;
  box-shadow: var(--primaryColor) 0 0 1em;
}

#listingView {
  min-height: 90vh !important;
}

.folder-items a {
  border-color: #d1d1d1;
  border-style: solid;
}

</style>
