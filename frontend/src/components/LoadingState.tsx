import React from "react";
import { View, Text, ActivityIndicator, StyleSheet } from "react-native";

export function LoadingState({ message }: { message?: string }) {
  return (
    <View style={styles.wrap} testID="loading-state">
      <ActivityIndicator size="large" testID="loading-spinner" />
      <Text style={styles.text}>{message ?? "Loading tasks..."}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { padding: 24, alignItems: "center", justifyContent: "center" },
  text: { marginTop: 8, color: "#666" },
});
