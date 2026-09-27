import React, { useEffect, useState } from "react";
import { TextInput, View, StyleSheet } from "react-native";

interface Props {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

/** Debounced search input — required Task 3 feature. */
export function SearchInput({ value, onChange, placeholder, debounceMs = 300 }: Props) {
  const [inner, setInner] = useState(value);

  useEffect(() => {
    setInner(value);
  }, [value]);

  useEffect(() => {
    const t = setTimeout(() => {
      if (inner !== value) onChange(inner);
    }, debounceMs);
    return () => clearTimeout(t);
  }, [inner, value, onChange, debounceMs]);

  return (
    <View style={styles.wrap}>
      <TextInput
        testID="search-input"
        style={styles.input}
        value={inner}
        onChangeText={setInner}
        placeholder={placeholder ?? "Search tasks..."}
        accessibilityLabel="Search tasks"
      />
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { marginVertical: 8 },
  input: {
    borderWidth: 1,
    borderColor: "#ccc",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    backgroundColor: "#fff",
  },
});
