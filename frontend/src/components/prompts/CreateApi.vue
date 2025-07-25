<template>
  <div class="card floating create-api__prompt__card" id="create-api">
    <div class="card-title">
      <h2>{{ $t('api.createTitle') }}</h2>
    </div>

    <div class="card-content">
      <!-- API Key Name Input -->
      <p>{{ $t('api.keyName') }}</p>
      <input
        class="input input--block"
        type="text"
        v-model.trim="apiName"
        :placeholder="$t('api.keyNamePlaceholder')"
      />

      <!-- Duration Input -->
      <p>{{ $t('api.tokenDuration') }}</p>
      <div class="inputWrapper">
        <input
          class="sizeInput roundedInputLeft input"
          v-model.number="duration"
          type="number"
          min="1"
          :placeholder="$t('api.durationNumberPlaceholder')"
        />
        <select v-model="unit" class="roundedInputRight input">
          <option value="days">{{ $t('api.days') }}</option>
          <option value="months">{{ $t('api.months') }}</option>
        </select>
      </div>

      <!-- Permissions Input -->
      <p>{{ $t('api.permissionNote') }}</p>
      <div class="settings-items">
        <ToggleSwitch
          v-for="(isEnabled, permission) in permissions"
          :key="permission"
          class="item"
          v-model="permissions[permission]"
          :name="permission"
        />
      </div>
    </div>

    <div class="card-action">
      <button
        @click="closeHovers"
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat button--blue"
        @click="createAPIKey"
        :title="$t('buttons.create')"
      >
        {{ $t("buttons.create") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mutations } from "@/store";
import { notify } from "@/notify";
import { usersApi } from "@/api";
import ToggleSwitch from "@/components/settings/ToggleSwitch.vue";

export default {
  name: "CreateAPI",
  data() {
    return {
      apiName: "",
      duration: 1,
      unit: "days",
    };
  },
  components: {
    ToggleSwitch,
  },
  props: {
    permissions: {
      type: Object,
      required: true,
    },
  },
  computed: {
    durationInDays() {
      // Calculate duration based on unit
      return this.unit === "days" ? this.duration : this.duration * 30; // assuming 30 days per month
    },
  },
  methods: {
    closeHovers() {
      mutations.closeHovers();
    },
    async createAPIKey() {
      try {
        // Filter to get keys of permissions set to true and join them as a comma-separated string
        const permissionsString = Object.keys(this.permissions)
          .filter((key) => this.permissions[key])
          .join(",");

        const params = {
          name: this.apiName,
          days: this.durationInDays,
          permissions: permissionsString,
        };

        await usersApi.createApiKey(params);
        notify.showSuccess($t("api.createKeySuccess"));
        window.location.reload();
        mutations.closeHovers();
      } catch (error) {
        notify.showError($t("api.createKeyFailed"));
      }
    },
  },
};
</script>
