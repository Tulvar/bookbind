type TaskHandlers<T> = {
    success: (result: T) => void;
    failure: (error: unknown) => void;
    finish: () => void;
};

export function createLatestTask() {
    let current = 0;
    return {
        invalidate() {
            current++;
        },
        async run<T>(load: () => Promise<T>, handlers: TaskHandlers<T>) {
            const request = ++current;
            try {
                const result = await load();
                if (request === current) {
                    handlers.success(result);
                }
            } catch (error) {
                if (request === current) {
                    handlers.failure(error);
                }
            } finally {
                if (request === current) {
                    handlers.finish();
                }
            }
        },
    };
}
