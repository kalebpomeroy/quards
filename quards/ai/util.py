import time
from functools import wraps


def timed(label):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            print(f"{label}...")
            start = time.time()
            result = func(*args, **kwargs)
            print(f"{label} completed in {time.time() - start:.3f}s")
            return result

        return wrapper

    return decorator
