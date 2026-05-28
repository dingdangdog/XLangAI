import { createError } from "h3";
import {
  isInternalSystemSettingKey,
  isKnownSystemSettingKey,
  KEY_META,
  MEDIA_STORAGE_VALUES,
  SYSTEM_SETTING_KEYS,
  VALUE_TYPES,
  type SystemSettingKey,
} from "./systemSettingsKeys";

export function assertEditableSystemSettingKey(key: string) {
  if (isInternalSystemSettingKey(key)) {
    throw createError({ statusCode: 403, message: "This setting is managed elsewhere" });
  }
}

function validateValue(key: SystemSettingKey, value: unknown, valueType: string) {
  const meta = KEY_META[key];
  const v = String(value ?? "").trim();
  if (valueType === "bool" || meta.valueType === "bool") {
    if (!["true", "false", "1", "0"].includes(v.toLowerCase())) {
      throw createError({ statusCode: 400, message: "bool value must be true or false" });
    }
    return v.toLowerCase() === "true" || v === "1" ? "true" : "false";
  }
  if (meta.enumValues && !meta.enumValues.includes(v as (typeof MEDIA_STORAGE_VALUES)[number])) {
    throw createError({
      statusCode: 400,
      message: `value must be one of: ${meta.enumValues.join(", ")}`,
    });
  }
  return v;
}

const STATUS_VALUES = ["active", "inactive"] as const;

export function prepareSystemSettingWrite(
  data: Record<string, unknown>,
  mode: "create" | "update",
  existingKey?: string,
) {
  const next = { ...data };
  if (typeof next.status !== "undefined") {
    const status = String(next.status).trim();
    if (!STATUS_VALUES.includes(status as (typeof STATUS_VALUES)[number])) {
      throw createError({ statusCode: 400, message: "invalid status" });
    }
    next.status = status;
  } else if (mode === "create") {
    next.status = "active";
  }
  const key = String(next.key ?? existingKey ?? "").trim();
  if (mode === "create") {
    if (!isKnownSystemSettingKey(key)) {
      throw createError({
        statusCode: 400,
        message: `unknown key; use one of: ${SYSTEM_SETTING_KEYS.join(", ")}`,
      });
    }
  } else {
    delete next.key;
  }
  const valueType = String(next.valueType ?? "string").trim();
  if (!VALUE_TYPES.includes(valueType as (typeof VALUE_TYPES)[number])) {
    throw createError({ statusCode: 400, message: "invalid value_type" });
  }
  if (mode === "create" && isKnownSystemSettingKey(key)) {
    const meta = KEY_META[key];
    if (!next.valueType) {
      next.valueType = meta.valueType;
    }
    next.value = validateValue(key, next.value, String(next.valueType));
  }
  if (mode === "update" && typeof next.value !== "undefined" && isKnownSystemSettingKey(key)) {
    const vt = String(next.valueType ?? KEY_META[key].valueType);
    next.value = validateValue(key, next.value, vt);
    if (!next.valueType) {
      next.valueType = KEY_META[key].valueType;
    }
  }
  return next;
}
