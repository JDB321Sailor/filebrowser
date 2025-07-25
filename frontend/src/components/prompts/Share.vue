<template>
  <div class="card floating share__promt__card" id="share">
    <div class="card-title">
      <h2>{{ $t("buttons.share") }}</h2>
    </div>
    <div aria-label="share-path" class="searchContext"> {{$t('search.path')}} {{ subpath }}</div>
    <p> {{ $t('share.notice') }} </p>
    <template v-if="listing">
      <div class="card-content">
        <table>
          <tbody>
            <tr>
              <th>#</th> <!-- eslint-disable-line @intlify/vue-i18n/no-raw-text -->
              <th>{{ $t("settings.shareDuration") }}</th>
              <th></th>
              <th></th>
            </tr>

            <tr v-for="link in links" :key="link.hash">
              <td>{{ link.hash }}</td>
              <td>
                <template v-if="link.expire !== 0">{{ humanTime(link.expire) }}</template>
                <template v-else>{{ $t("permanent") }}</template>
              </td>
              <td class="small">
                <button
                  class="action copy-clipboard"
                  :data-clipboard-text="buildLink(link)"
                  :aria-label="$t('buttons.copyToClipboard')"
                  :title="$t('buttons.copyToClipboard')"
                >
                  <i class="material-icons">content_paste</i>
                </button>
              </td>
              <td class="small" v-if="hasDownloadLink()">
                <button
                  class="action copy-clipboard"
                  :data-clipboard-text="buildDownloadLink(link)"
                  :aria-label="$t('buttons.copyDownloadLinkToClipboard')"
                  :title="$t('buttons.copyDownloadLinkToClipboard')"
                >
                  <i class="material-icons">content_paste_go</i>
                </button>
              </td>
              <td class="small">
                <button
                  class="action"
                  @click="deleteLink($event, link)"
                  :aria-label="$t('buttons.delete')"
                  :title="$t('buttons.delete')"
                >
                  <i class="material-icons">delete</i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.close')"
          :title="$t('buttons.close')"
        >
          {{ $t("buttons.close") }}
        </button>
        <button
          class="button button--flat button--blue"
          @click="() => switchListing()"
          :aria-label="$t('buttons.new')"
          :title="$t('buttons.new')"
        >
          {{ $t("buttons.new") }}
        </button>
      </div>
    </template>

    <template v-else>
      <div class="card-content">
        <p>{{ $t("settings.shareDuration") }}</p>
        <div class="input-group input">
          <input
            v-focus
            type="number"
            max="2147483647"
            min="1"
            @keyup.enter="submit"
            v-model.trim="time"
          />
          <select class="right" v-model="unit" :aria-label="$t('time.unit')">
            <option value="seconds">{{ $t("time.seconds") }}</option>
            <option value="minutes">{{ $t("time.minutes") }}</option>
            <option value="hours">{{ $t("time.hours") }}</option>
            <option value="days">{{ $t("time.days") }}</option>
          </select>
        </div>
        <p>{{ $t("prompts.optionalPassword") }}</p>
        <input class="input input--block" type="password" v-model.trim="password" />
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="() => switchListing()"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          class="button button--flat button--blue"
          @click="submit"
          aria-label="Share-Confirm"
          :title="$t('buttons.share')"
        >
          {{ $t("buttons.share") }}
        </button>
      </div>
    </template>
  </div>
</template>
<script>
import { notify } from "@/notify";
import { state, getters, mutations } from "@/store";
import { shareApi, publicApi } from "@/api";
import Clipboard from "clipboard";
import { fromNow } from "@/utils/moment";
import { buildItemUrl } from "@/utils/url";

export default {
  name: "share",
  data() {
    return {
      time: "",
      unit: "hours",
      links: [],
      clip: null,
      subpath: "",
      source: "",
      password: "",
      listing: true,
    };
  },
  computed: {
    closeHovers() {
      return mutations.closeHovers;
    },
    req() {
      return state.req; // Access state directly
    },
    selected() {
      return state.selected; // Access state directly
    },
    selectedCount() {
      return state.selected.length; // Compute selectedCount directly from state
    },
    isListing() {
      return getters.isListing(); // Access getter directly from the store
    },
    url() {
      if (state.isSearchActive) {
        return buildItemUrl(state.selected[0].source,state.selected[0].path)
      }
      if (!this.isListing) {
        return state.route.path;
      }
      if (getters.selectedCount() !== 1) {
        // selecting current view image
        return state.route.path;
      }
      return buildItemUrl(state.req.items[this.selected[0]].source,state.req.items[this.selected[0]].path)
    },
  },
  async beforeMount() {
    let path = state.req.path;
    this.source = state.req.source;
    if (state.isSearchActive) {
      path = state.selected[0].path;
      this.source = state.selected[0].source;
    } else if (getters.selectedCount() === 1) {
      const selected = getters.getFirstSelected();
      path = selected.path;
      this.source = selected.source;
      this.source = state.req.items[state.selected[0]].source;
    }
    // double encode # to fix issue with # in path
    // replace all # with %23
    this.subpath = path.replace(/#/g, "%23");

    try {
      // get last element of the path
      const links = await shareApi.get(this.subpath, this.source);
      this.links = links;
    } catch (err) {
      notify.showError(err);
      return;
    }
    this.sort();

    if (this.links.length === 0) {
      this.listing = false;
    }
  },
  mounted() {
    this.clip = new Clipboard(".copy-clipboard");
    this.clip.on("success", () => {
      notify.showSuccess(this.$t("success.linkCopied"));
    });
  },
  beforeUnmount() {
    this.clip.destroy();
  },
  methods: {
    async submit() {
      try {
        let isPermanent = !this.time || this.time == 0;
        let res = null;
        if (isPermanent) {
          res = await shareApi.create(this.subpath, this.source, this.password);
        } else {
          res = await shareApi.create(
            this.subpath,
            this.source,
            this.password,
            this.time.toString(),
            this.unit
          );
        }

        this.links.push(res);
        this.sort();

        this.time = "";
        this.unit = "hours";
        this.password = "";

        this.listing = true;
      } catch (err) {
        notify.showError(err);
      }
    },
    async deleteLink(event, link) {
      event.preventDefault();
      try {
        await shareApi.remove(link.hash);
        this.links = this.links.filter((item) => item.hash !== link.hash);
        if (this.links.length === 0) {
          this.listing = false;
        }
      } catch (err) {
        notify.showError(err);
      }
    },
    humanTime(time) {
      return fromNow(time, state.user.locale)
    },
    buildLink(share) {
      return shareApi.getShareURL(share);
    },
    hasDownloadLink() {
      if (state.isSearchActive) {
        return state.selected[0].type != "directory";
      }
      return this.selected.length === 1 && !state.req.items[this.selected[0]].isDir;
    },
    buildDownloadLink(share) {
      share.source = this.source;
      share.path = "/";
      return publicApi.getDownloadURL(share);
    },
    sort() {
      this.links = this.links.sort((a, b) => {
        if (a.expire === 0) return -1;
        if (b.expire === 0) return 1;
        return new Date(a.expire) - new Date(b.expire);
      });
    },
    switchListing() {
      if (this.links.length === 0 && !this.listing) {
        // Access the store directly if needed
        mutations.closeHovers();
      }

      this.listing = !this.listing;
    },
  },
};
</script>
