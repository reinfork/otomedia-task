import React, { useEffect, useState } from "react";
import {
  Modal,
  View,
  Text,
  TextInput,
  Pressable,
  StyleSheet,
  ActivityIndicator,
} from "react-native";
import type { Task } from "../api/client";

interface Props {
  visible: boolean;
  task: Task | null;
  saving: boolean;
  error: string | null;
  onClose: () => void;
  onSave: (payload: { title: string; description: string; status: Task["status"]; assignee: string }) => void;
}

/** Edit modal — required Task 3 feature (PUT /api/tasks/:id). */
export function EditModal({ visible, task, saving, error, onClose, onSave }: Props) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [assignee, setAssignee] = useState("");
  const [status, setStatus] = useState<Task["status"]>("todo");

  useEffect(() => {
    if (task) {
      setTitle(task.title);
      setDescription(task.description ?? "");
      setAssignee(task.assignee ?? "");
      setStatus(task.status);
    }
  }, [task]);

  return (
    <Modal visible={visible} animationType="slide" transparent onRequestClose={onClose}>
      <View style={styles.backdrop}>
        <View style={styles.sheet}>
          <Text style={styles.heading}>Edit task</Text>
          <TextInput
            testID="edit-title"
            style={styles.input}
            value={title}
            onChangeText={setTitle}
            placeholder="Title"
          />
          <TextInput
            testID="edit-description"
            style={styles.input}
            value={description}
            onChangeText={setDescription}
            placeholder="Description"
          />
          <TextInput
            testID="edit-assignee"
            style={styles.input}
            value={assignee}
            onChangeText={setAssignee}
            placeholder="Assignee"
          />
          <View style={styles.statusRow}>
            {(["todo", "in_progress", "done"] as const).map((s) => (
              <Pressable
                key={s}
                testID={`edit-status-${s}`}
                onPress={() => setStatus(s)}
                style={[styles.chip, status === s && styles.chipActive]}
              >
                <Text style={status === s ? styles.chipTextActive : undefined}>{s}</Text>
              </Pressable>
            ))}
          </View>
          {error ? (
            <Text testID="edit-error" style={styles.error}>
              {error}
            </Text>
          ) : null}
          <View style={styles.actions}>
            <Pressable testID="edit-cancel" onPress={onClose} style={styles.btn}>
              <Text>Cancel</Text>
            </Pressable>
            <Pressable
              testID="edit-save"
              disabled={saving || title.trim() === ""}
              onPress={() => onSave({ title: title.trim(), description, status, assignee })}
              style={[styles.btn, styles.primary]}
            >
              {saving ? <ActivityIndicator /> : <Text style={styles.primaryText}>Save</Text>}
            </Pressable>
          </View>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: "rgba(0,0,0,0.4)",
    justifyContent: "flex-end",
  },
  sheet: { backgroundColor: "#fff", padding: 16, borderTopLeftRadius: 16, borderTopRightRadius: 16 },
  heading: { fontSize: 18, fontWeight: "700", marginBottom: 12 },
  input: {
    borderWidth: 1,
    borderColor: "#ccc",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    marginBottom: 8,
  },
  statusRow: { flexDirection: "row", gap: 8, marginVertical: 8 },
  chip: { paddingHorizontal: 10, paddingVertical: 6, borderWidth: 1, borderColor: "#aaa", borderRadius: 14 },
  chipActive: { backgroundColor: "#111", borderColor: "#111" },
  chipTextActive: { color: "#fff" },
  error: { color: "red", marginVertical: 6 },
  actions: { flexDirection: "row", justifyContent: "flex-end", gap: 8, marginTop: 8 },
  btn: { paddingHorizontal: 14, paddingVertical: 10, borderWidth: 1, borderColor: "#aaa", borderRadius: 8 },
  primary: { backgroundColor: "#111", borderColor: "#111" },
  primaryText: { color: "#fff" },
});
