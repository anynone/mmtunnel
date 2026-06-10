import { Select } from "./ui/select.jsx";

export function ThemeModeControl({ value, onChange }) {
  return (
    <Select value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="system">System</option>
      <option value="light">Light</option>
      <option value="dark">Dark</option>
    </Select>
  );
}
