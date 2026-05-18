import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { TagSelector } from './TagSelector';
import { createTestWrapper } from '../test/wrapper';

const mockCreateMutate = vi.fn();

vi.mock('../hooks/useTags', () => ({
  useTags: () => ({
    data: [
      { id: 'tag-1', name: 'production', color: '#f50' },
      { id: 'tag-2', name: 'critical', color: '#2db7f5' },
      { id: 'tag-3', name: 'staging', color: '#87d068' },
    ],
  }),
  useCreateTag: () => ({ mutate: mockCreateMutate, isPending: false }),
}));

describe('TagSelector', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders assigned tags by ID lookup', () => {
    render(<TagSelector value={['tag-1', 'tag-2']} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    expect(screen.getByText('production')).toBeInTheDocument();
    expect(screen.getByText('critical')).toBeInTheDocument();
  });

  it('shows empty state when no tags assigned', () => {
    render(<TagSelector value={[]} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    expect(screen.getByText('No tags assigned')).toBeInTheDocument();
  });

  it('calls onChange without the removed tag when close is clicked', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagSelector value={['tag-1', 'tag-2']} onChange={onChange} />, {
      wrapper: createTestWrapper(),
    });

    const closeIcons = document.querySelectorAll('.ant-tag .anticon-close');
    await user.click(closeIcons[0] as Element);

    expect(onChange).toHaveBeenCalledWith(['tag-2']);
  });

  it('calls onChange with added tag when a tag is selected from dropdown', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagSelector value={['tag-1', 'tag-2']} onChange={onChange} />, {
      wrapper: createTestWrapper(),
    });

    const combobox = screen.getByRole('combobox');
    await user.click(combobox);
    await user.type(combobox, 'staging');

    const option = await screen.findByText('staging');
    await user.click(option);

    expect(onChange).toHaveBeenCalledWith(['tag-1', 'tag-2', 'tag-3']);
  });

  it('shows create form when "New Tag" is clicked', async () => {
    const user = userEvent.setup();
    render(<TagSelector value={[]} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole('button', { name: /new tag/i }));

    expect(screen.getByPlaceholderText('Tag name')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Add' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
  });

  it('calls createTag when new tag form is submitted', async () => {
    const user = userEvent.setup();
    render(<TagSelector value={[]} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole('button', { name: /new tag/i }));
    await user.type(screen.getByPlaceholderText('Tag name'), 'deployment');
    await user.click(screen.getByRole('button', { name: 'Add' }));

    expect(mockCreateMutate).toHaveBeenCalledWith(
      { name: 'deployment', color: '#597ef7' },
      expect.any(Object),
    );
  });

  it('hides create form when cancel is clicked', async () => {
    const user = userEvent.setup();
    render(<TagSelector value={[]} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole('button', { name: /new tag/i }));
    expect(screen.getByPlaceholderText('Tag name')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByPlaceholderText('Tag name')).not.toBeInTheDocument();
  });

  it('disables Add button when tag name is empty', async () => {
    const user = userEvent.setup();
    render(<TagSelector value={[]} onChange={() => {}} />, {
      wrapper: createTestWrapper(),
    });

    await user.click(screen.getByRole('button', { name: /new tag/i }));
    expect(screen.getByRole('button', { name: 'Add' })).toBeDisabled();
  });
});
