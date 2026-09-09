import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & {
	ref?: U | null;
};

export type WithoutChildren<T> = T extends { children?: unknown } ? Omit<T, 'children'> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<Omit<T, 'child'>>;
export type WithoutChild<T> = Omit<T, 'child'>;
