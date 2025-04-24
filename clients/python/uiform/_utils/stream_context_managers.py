from contextlib import AbstractAsyncContextManager, AbstractContextManager
from typing import Any, AsyncGenerator, Callable, Generator, TypeVar, Union

T = TypeVar('T')


class AsyncGeneratorContextManager(AbstractAsyncContextManager[AsyncGenerator[T, None]]):
    def __init__(self, generator_func: Callable[..., AsyncGenerator[T, None]], *args: Any, **kwargs: Any):
        self.generator_func = generator_func
        self.args = args
        self.kwargs = kwargs
        self.iterator: Union[AsyncGenerator[T, None], None] = None

    async def __aenter__(self) -> AsyncGenerator[T, None]:
        # Create the asynchronous iterator from the generator function
        self.iterator = self.generator_func(*self.args, **self.kwargs)
        return self.iterator

    async def __aexit__(self, exc_type: type[BaseException] | None, exc_value: BaseException | None, traceback: Any) -> None:
        # Ensure the iterator is properly closed if it supports aclose
        if self.iterator is not None:
            await self.iterator.aclose()


class GeneratorContextManager(AbstractContextManager[Generator[T, None, None]]):
    def __init__(self, generator_func: Callable[..., Generator[T, None, None]], *args: Any, **kwargs: Any):
        self.generator_func = generator_func
        self.args = args
        self.kwargs = kwargs
        self.iterator: Union[Generator[T, None, None], None] = None

    def __enter__(self) -> Generator[T, None, None]:
        self.iterator = self.generator_func(*self.args, **self.kwargs)
        return self.iterator

    def __exit__(self, exc_type: type[BaseException] | None, exc_value: BaseException | None, traceback: Any) -> None:
        if self.iterator is not None:
            self.iterator.close()


def as_async_context_manager(func: Callable[..., AsyncGenerator[T, None]]) -> Callable[..., AsyncGeneratorContextManager[T]]:
    def wrapper(*args: Any, **kwargs: Any) -> AsyncGeneratorContextManager[T]:
        return AsyncGeneratorContextManager(func, *args, **kwargs)

    return wrapper


def as_context_manager(func: Callable[..., Generator[T, None, None]]) -> Callable[..., GeneratorContextManager[T]]:
    def wrapper(*args: Any, **kwargs: Any) -> GeneratorContextManager[T]:
        return GeneratorContextManager(func, *args, **kwargs)

    return wrapper
