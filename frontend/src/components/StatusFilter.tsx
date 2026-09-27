import React from "react";
import { View, Pressable, Text, StyleSheet } from "react-native";

export const STATUS_OPTIONS = ["all", "todo", "in_progress", "done"] as const;
export type StatusOption = (typeof STATUS_OPTIONS)[number];

interface Props {
  value: StatusOption;
  onChange: (v: StatusOption) => void;
}

export function StatusFilter({ value, onChange }: Props) {
  return (
    <View style={styles.row} testID="status-filter">
      {STATUS_OPTIONS.map((s) => {
        const active = value === s;
        return (
          <Pressable
            key={s}
            testID={`status-${s}`}
            accessibilityRole="button"
            onPress={() => onChange(s)}
            style={[styles.chip, active && styles.chipActive]}
          >
            <Text style={[styles.label, active && styles.labelActive]}>{s}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: "row", flexWrap: "wrap", gap: 8, marginVertical: 8 },
  chip: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: "#aaa",
  },
  chipActive: { backgroundColor: "#111", borderColor: "#111" },
  label: { fontSize: 14, color: "#333" },
  labelActive: { color: "#fff" },
});
