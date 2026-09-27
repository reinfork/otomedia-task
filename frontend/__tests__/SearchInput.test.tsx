import React from "react";
import { render, fireEvent, waitFor } from "@testing-library/react-native";
import { SearchInput } from "../src/components/SearchInput";

jest.useFakeTimers();

describe("SearchInput", () => {
  it("renders and debounces onChange", async () => {
    const onChange = jest.fn();
    const { getByTestId } = render(<SearchInput value="" onChange={onChange} />);

    const input = getByTestId("search-input");
    expect(input).toBeTruthy();

    fireEvent.changeText(input, "fix");
    // Not called synchronously due to debounce.
    expect(onChange).not.toHaveBeenCalled();

    jest.advanceTimersByTime(400);
    await waitFor(() => expect(onChange).toHaveBeenCalledWith("fix"));
  });

  it("shows placeholder", () => {
    const { getByPlaceholderText } = render(
      <SearchInput value="" onChange={() => {}} placeholder="Search tasks..." />
    );
    expect(getByPlaceholderText("Search tasks...")).toBeTruthy();
  });
});
