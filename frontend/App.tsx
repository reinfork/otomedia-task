import React, { useMemo, useState } from "react";
import {
  SafeAreaView,
  View,
  Text,
  FlatList,
  StyleSheet,
  Alert,
  RefreshControl,
} from "react-native";
import { StatusBar } from "expo-status-bar";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SearchInput } from "./src/components/SearchInput";
import { StatusFilter, StatusOption } from "./src/components/StatusFilter";
import { Pagination } from "./src/components/Pagination";
import { TaskCard } from "./src/components/TaskCard";
import { EditModal } from "./src/components/EditModal";
import { LoadingState } from "./src/components/LoadingState";
import { useTasks, useUpdateTask, useDeleteTask } from "./src/hooks/useTasks";
import type { Task } from "./src/api/client";

const queryClient = new QueryClient();

function TaskListScreen() {
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState<StatusOption>("all");
  const [page, setPage] = useState(1);
  const [editing, setEditing] = useState<Task | null>(null);
  const limit = 10;

  const params = useMemo(
    () => ({
      keyword: keyword || undefined,
      status: status === "all" ? undefined : status,
      page,
      limit,
      sort: "-created_at",
    }),
    [keyword, status, page]
  );

  const { data, isLoading, isError, refetch, isRefetching } = useTasks(params);
  const updateMutation = useUpdateTask();
  const deleteMutation = useDeleteTask();

  const tasks = data?.data ?? [];
  const totalPages = data?.meta.total_pages ?? 0;

  const handleFilterChange = (fn: () => void) => {
    fn();
    setPage(1);
  };

  return (
    <SafeAreaView style={styles.container}>
      <Text style={styles.heading}>Tasks</Text>
      <SearchInput value={keyword} onChange={(v) => handleFilterChange(() => setKeyword(v))} />
      <StatusFilter value={status} onChange={(v) => handleFilterChange(() => setStatus(v))} />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <View testID="error-state" style={styles.center}>
          <Text>Failed to load tasks.</Text>
          <Text style={styles.retry} onPress={() => refetch()}>
            Tap to retry
          </Text>
        </View>
      ) : tasks.length === 0 ? (
        <View testID="empty-state" style={styles.center}>
          <Text>No tasks found.</Text>
        </View>
      ) : (
        <FlatList
          testID="task-list"
          data={tasks}
          keyExtractor={(t) => t.id}
          renderItem={({ item }) => (
            <TaskCard
              task={item}
              onEdit={setEditing}
              onDelete={(t) =>
                Alert.alert("Delete task", `Delete "${t.title}"?`, [
                  { text: "Cancel", style: "cancel" },
                  {
                    text: "Delete",
                    style: "destructive",
                    onPress: () => deleteMutation.mutate(t.id),
                  },
                ])
              }
            />
          )}
          refreshControl={
            <RefreshControl refreshing={isRefetching} onRefresh={() => refetch()} />
          }
        />
      )}

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />

      <EditModal
        visible={editing !== null}
        task={editing}
        saving={updateMutation.isPending}
        error={updateMutation.isError ? "Failed to save. Title may already exist." : null}
        onClose={() => {
          setEditing(null);
          updateMutation.reset();
        }}
        onSave={(payload) => {
          if (!editing) return;
          updateMutation.mutate(
            { id: editing.id, payload },
            { onSuccess: () => setEditing(null) }
          );
        }}
      />
      <StatusBar style="auto" />
    </SafeAreaView>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <TaskListScreen />
    </QueryClientProvider>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, padding: 16, backgroundColor: "#f6f6f6" },
  heading: { fontSize: 24, fontWeight: "800", marginBottom: 4 },
  center: { padding: 24, alignItems: "center" },
  retry: { color: "blue", marginTop: 8 },
});
