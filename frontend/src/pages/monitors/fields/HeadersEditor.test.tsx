import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { HeadersEditor } from './HeadersEditor';

describe('HeadersEditor', () => {
  it('renders empty state with add button', () => {
    render(<HeadersEditor value={[]} onChange={() => {}} />);
    expect(screen.getByRole('button', { name: /add header/i })).toBeInTheDocument();
  });

  it('renders existing entries', () => {
    render(
      <HeadersEditor
        value={[
          { name: 'Authorization', value: 'Bearer token' },
          { name: 'Accept', value: 'application/json' },
        ]}
        onChange={() => {}}
      />,
    );

    const inputs = screen.getAllByPlaceholderText('Value');
    expect(inputs).toHaveLength(2);
    expect(inputs[0]).toHaveValue('Bearer token');
    expect(inputs[1]).toHaveValue('application/json');
  });

  it('calls onChange with new entry when Add Header is clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<HeadersEditor value={[]} onChange={onChange} />);

    await user.click(screen.getByRole('button', { name: /add header/i }));

    // Empty-name entries are filtered out of onChange
    expect(onChange).toHaveBeenCalledWith([]);
  });

  it('calls onChange with updated array when header name is typed', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<HeadersEditor value={[]} onChange={onChange} />);

    await user.click(screen.getByRole('button', { name: /add header/i }));

    const nameInput = screen.getByRole('combobox');
    await user.type(nameInput, 'X-Custom');

    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1][0];
    expect(lastCall).toEqual([{ name: 'X-Custom', value: '' }]);
  });

  it('removes entry when delete button is clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <HeadersEditor
        value={[
          { name: 'Keep', value: 'yes' },
          { name: 'Remove', value: 'no' },
        ]}
        onChange={onChange}
      />,
    );

    const removeButtons = screen.getAllByRole('button', { name: /remove header/i });
    await user.click(removeButtons[1]);

    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1][0];
    expect(lastCall).toEqual([{ name: 'Keep', value: 'yes' }]);
  });

  it('emits value with both name and value populated', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<HeadersEditor value={[]} onChange={onChange} />);

    await user.click(screen.getByRole('button', { name: /add header/i }));

    const nameInput = screen.getByRole('combobox');
    const valueInput = screen.getByPlaceholderText('Value');

    await user.type(nameInput, 'Content-Type');
    await user.type(valueInput, 'text/plain');

    const lastCall = onChange.mock.calls[onChange.mock.calls.length - 1][0];
    expect(lastCall).toEqual([{ name: 'Content-Type', value: 'text/plain' }]);
  });
});
