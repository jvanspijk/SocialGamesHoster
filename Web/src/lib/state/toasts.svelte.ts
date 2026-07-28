export type ToastTone = 'error' | 'success' | 'info';

export interface ToastMessage {
	id: string;
	tone: ToastTone;
	message: string;
	actionLabel?: string;
	action?: () => void;
	persistent?: boolean;
}

let messages = $state<ToastMessage[]>([]);
let nextId = 0;

function add(
	tone: ToastTone,
	message: string,
	options: Pick<ToastMessage, 'actionLabel' | 'action' | 'persistent'> = {}
) {
	const duplicate = messages.find((item) => item.tone === tone && item.message === message);
	if (duplicate) return duplicate.id;
	const id = `toast-${++nextId}`;
	messages = [...messages.slice(-2), { id, tone, message, ...options }];
	return id;
}

export const toasts = {
	get items() {
		return messages;
	},
	error(message: string, options?: Pick<ToastMessage, 'actionLabel' | 'action' | 'persistent'>) {
		return add('error', message, options);
	},
	success(message: string) {
		return add('success', message);
	},
	info(message: string, options?: Pick<ToastMessage, 'actionLabel' | 'action' | 'persistent'>) {
		return add('info', message, options);
	},
	dismiss(id: string) {
		messages = messages.filter((item) => item.id !== id);
	},
	clear() {
		messages = [];
	}
};
