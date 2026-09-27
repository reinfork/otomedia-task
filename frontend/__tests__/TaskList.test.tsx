import React from "react";
import { render, fireEvent } from "@testing-library/react-native";
import { TaskCard } from "../src/components/TaskCard";
import { Pagination } from "../src/components/Pagination";

const task = {
  id: "11111111-1111-1111-1111-111111111111",
  title: "Fix login bug",
  description: "auth fails on retry",
  status: "todo" as const,
  assignee: "budi",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe("Task list components", () => {
  it("renders task card and fires edit", () => {
    const onEdit = jest.fn();
    const { getByText, getByTestId } = render(
      <TaskCard task={task} onEdit={onEdit} onDelete={() => {}} />
    );
    expect(getByText("Fix login bug")).toBeTruthy();
    fireEvent.press(getByTestId(`edit-${task.id}`));
    expect(onEdit).toHaveBeenCalledWith(task);
  });

  it("paginates forward and back", () => {
    const onPageChange = jest.fn();
    const { getByTestId } = render(
      <Pagination page={2} totalPages={5} onPageChange={onPageChange} />
    );
    fireEvent.press(getByTestId("next-page"));
    expect(onPageChange).toHaveBeenCalledWith(3);
    fireEvent.press(getByTestId("prev-page"));
    expect(onPageChange).toHaveBeenCalledWith(1);
  });
});
