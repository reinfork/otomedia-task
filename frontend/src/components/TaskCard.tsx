import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import type { Task } from "../api/client";

interface Props {
  task: Task;
  onEdit: (t: Task) => void;
  onDelete: (t: Task) => void;
}

export function TaskCard({ task, onEdit, onDelete }: Props) {
  return (
    <View style={styles.card} testID={`task-${task.id}`}>
      <Text style={styles.title}>{task.title}</Text>
      {!!task.description && <Text style={styles.desc}>{task.description}</Text>}
      <Text style={styles.meta}>
        {task.status} • {task.assignee || "unassigned"}
      </Text>
      <View style={styles.actions}>
        <Pressable testID={`edit-${task.id}`} onPress={() => onEdit(task)} style={styles.btn}>
          <Text>Edit</Text>
        </Pressable>
        <Pressable testID={`delete-${task.id}`} onPress={() => onDelete(task)} style={styles.btn}>
          <Text>Delete</Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderWidth: 1,
    borderColor: "#ddd",
    borderRadius: 10,
    padding: 12,
    marginVertical: 6,
    backgroundColor: "#fff",
  },
  title: { fontSize: 16, fontWeight: "600" },
  desc: { fontSize: 14, color: "#555", marginTop: 4 },
  meta: { fontSize: 12, color: "#888", marginTop: 6 },
  actions: { flexDirection: "row", gap: 8, marginTop: 8 },
  btn: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderWidth: 1,
    borderColor: "#aaa",
    borderRadius: 6,
  },
});
