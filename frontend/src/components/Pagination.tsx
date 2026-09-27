import React from "react";
import { View, Text, Pressable, StyleSheet } from "react-native";

interface Props {
  page: number;
  totalPages: number;
  onPageChange: (p: number) => void;
}

export function Pagination({ page, totalPages, onPageChange }: Props) {
  if (totalPages <= 1) return null;
  return (
    <View style={styles.row} testID="pagination">
      <Pressable
        testID="prev-page"
        disabled={page <= 1}
        onPress={() => onPageChange(page - 1)}
        style={[styles.btn, page <= 1 && styles.disabled]}
      >
        <Text>Prev</Text>
      </Pressable>
      <Text testID="page-info">
        Page {page} / {totalPages}
      </Text>
      <Pressable
        testID="next-page"
        disabled={page >= totalPages}
        onPress={() => onPageChange(page + 1)}
        style={[styles.btn, page >= totalPages && styles.disabled]}
      >
        <Text>Next</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginVertical: 12,
  },
  btn: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderWidth: 1,
    borderColor: "#aaa",
    borderRadius: 8,
  },
  disabled: { opacity: 0.4 },
});
