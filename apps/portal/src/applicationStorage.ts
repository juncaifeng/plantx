export const SELECTED_APP_STORAGE_KEY = 'plantx-selected-application';

export function getStoredApplicationId(): string | null {
  return typeof window !== 'undefined'
    ? localStorage.getItem(SELECTED_APP_STORAGE_KEY)
    : null;
}

export function storeApplicationId(id: string | null): void {
  if (typeof window === 'undefined') return;
  if (id) {
    localStorage.setItem(SELECTED_APP_STORAGE_KEY, id);
  } else {
    localStorage.removeItem(SELECTED_APP_STORAGE_KEY);
  }
}
