<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.rename") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.renameMessage") }} <code>{{ oldName() }}</code>:</p> <!-- eslint-disable-line @intlify/vue-i18n/no-raw-text -->
      <input
        class="input input--block"
        v-focus
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button busubmittton--flat"
        type="submit"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')"
      >
        {{ $t("buttons.rename") }}
      </button>
    </div>
  </div>
</template>
<script>
import { filesApi } from "@/api";
import { state, getters, mutations } from "@/store";
import { notify } from "@/notify";

export default {
  name: "rename",
  data() {
    return {
      name: "",
    };
  },
  created() {
    this.name = this.oldName();
  },
  computed: {
    req() {
      return state.req;
    },
    selected() {
      return state.selected;
    },
    selectedCount() {
      return state.selectedCount;
    },
    isListing() {
      return getters.isListing();
    },
    closeHovers() {
      return mutations.closeHovers;
    },
  },
  methods: {
    cancel() {
      mutations.closeHovers();
    },
    oldName() {
      if (!this.isListing) {
        return state.req.name;
      }

      if (getters.selectedCount() === 0 || getters.selectedCount() > 1) {
        return;
      }

      return state.req.items[this.selected[0]].name;
    },
    async submit() {
      try {
        const items = [{
          from: state.req.path + "/" + state.req.items[this.selected[0]].name,
          fromSource: state.req.source,
          to: state.req.path + "/" + this.name,
          toSource: state.req.source,
        }];

        await filesApi.moveCopy(items, "move");
        mutations.closeHovers();
        if (!this.isListing) {
          this.$router.push({ path: newLink });
          return;
        }
      } catch (error) {
        notify.showError(error);
      }
    },
  },
};
</script>
