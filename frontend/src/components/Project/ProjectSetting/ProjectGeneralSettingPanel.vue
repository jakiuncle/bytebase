<template>
  <form class="max-w-md space-y-4">
    <p class="text-lg font-medium leading-7 text-main">
      {{ $t("common.general") }}
    </p>
    <div class="flex justify-between">
      <dl class="">
        <dt class="text-sm font-medium text-control-light">
          {{ $t("common.name") }} <span class="text-red-600">*</span>
        </dt>
        <dd class="mt-1 text-sm text-main">
          <input
            id="projectName"
            v-model="state.name"
            :disabled="!allowEdit"
            required
            autocomplete="off"
            type="text"
            class="textfield"
          />
        </dd>
        <ResourceIdField
          resource-type="project"
          :value="project.resourceId"
          :readonly="true"
        />
      </dl>

      <dl class="">
        <dt class="text-sm font-medium text-control-light">
          {{ $t("common.key") }} <span class="text-red-600">*</span>
        </dt>
        <dd class="mt-1 text-sm text-main">
          <input
            id="projectKey"
            v-model="state.key"
            :disabled="!allowEdit"
            required
            autocomplete="off"
            type="text"
            class="textfield uppercase"
          />
        </dd>
      </dl>
    </div>

    <div class="flex flex-col">
      <div for="name" class="text-sm font-medium text-control-light">
        {{ $t("common.mode") }}
        <span class="text-red-600">*</span>
      </div>
      <div class="mt-2 textlabel">
        <div class="radio-set-row">
          <label class="radio">
            <input
              v-model="state.tenantMode"
              tabindex="-1"
              type="radio"
              class="btn"
              value="DISABLED"
            />
            <span class="label">{{ $t("project.mode.standard") }}</span>
          </label>
          <label class="radio">
            <input
              v-model="state.tenantMode"
              tabindex="-1"
              type="radio"
              class="btn"
              value="TENANT"
            />
            <span class="label">{{ $t("project.mode.tenant") }}</span>
            <FeatureBadge
              feature="bb.feature.multi-tenancy"
              class="text-accent"
            />
          </label>
        </div>
      </div>
    </div>

    <div v-if="allowEdit" class="flex justify-end">
      <button
        type="button"
        class="btn-primary"
        :disabled="!allowSave"
        @click.prevent="save"
      >
        {{ $t("common.save") }}
      </button>
    </div>

    <FeatureModal
      v-if="state.requiredFeature"
      :feature="state.requiredFeature"
      @cancel="state.requiredFeature = undefined"
    />
  </form>
</template>

<script lang="ts">
import isEmpty from "lodash-es/isEmpty";
import { computed, defineComponent, PropType, reactive } from "vue";
import { useI18n } from "vue-i18n";
import {
  DEFAULT_PROJECT_ID,
  Project,
  ProjectPatch,
  ProjectTenantMode,
  FeatureType,
} from "../../../types";
import { hasFeature, pushNotification, useProjectStore } from "@/store";
import FeatureModal from "@/components/FeatureModal.vue";
import ResourceIdField from "@/components/v2/Form/ResourceIdField.vue";

interface LocalState {
  name: string;
  key: string;
  tenantMode: ProjectTenantMode;
  requiredFeature: FeatureType | undefined;
}

export default defineComponent({
  name: "ProjectGeneralSettingPanel",
  components: {
    FeatureModal,
    ResourceIdField,
  },
  props: {
    project: {
      required: true,
      type: Object as PropType<Project>,
    },
    allowEdit: {
      default: true,
      type: Boolean,
    },
  },
  setup(props) {
    const { t } = useI18n();
    const projectStore = useProjectStore();

    const state = reactive<LocalState>({
      name: props.project.name,
      key: props.project.key,
      tenantMode: props.project.tenantMode,
      requiredFeature: undefined,
    });

    const allowSave = computed((): boolean => {
      return (
        props.project.id != DEFAULT_PROJECT_ID &&
        !isEmpty(state.name) &&
        (state.name !== props.project.name ||
          state.key !== props.project.key ||
          state.tenantMode !== props.project.tenantMode)
      );
    });

    const save = () => {
      const projectPatch: ProjectPatch = {};

      if (state.name !== props.project.name) {
        projectPatch.name = state.name;
      }
      if (state.key !== props.project.key) {
        projectPatch.key = state.key;
      }
      if (state.tenantMode !== props.project.tenantMode) {
        if (state.tenantMode === "TENANT") {
          if (!hasFeature("bb.feature.multi-tenancy")) {
            state.tenantMode = "DISABLED";
            state.requiredFeature = "bb.feature.multi-tenancy";
            return;
          }
        }
        projectPatch.tenantMode = state.tenantMode;
      }

      projectStore
        .patchProject({
          projectId: props.project.id,
          projectPatch,
        })
        .then((updatedProject: Project) => {
          pushNotification({
            module: "bytebase",
            style: "SUCCESS",
            title: t("project.settings.success-updated"),
          });
          state.name = updatedProject.name;
          state.key = updatedProject.key;
          state.tenantMode = updatedProject.tenantMode;
        });
    };

    return {
      state,
      allowSave,
      save,
    };
  },
});
</script>
