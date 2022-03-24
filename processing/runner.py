import concurrent.futures
from typing import Callable
from uuid import uuid4

from PyQt6.QtCore import QObject, QRunnable, pyqtSignal, pyqtSlot


class RunnerSignals(QObject):
    """
    Defines the signals available from a running worker thread.

    progress
        int progress complete,from 0-100

    """

    results = pyqtSignal(object)
    progress = pyqtSignal(str, int)
    finished = pyqtSignal(bool)


class Runner(QRunnable):
    """Runner thread. Inherits from QRunnable to handle runner thread setup, signals and wrap-up."""

    def __init__(self, job_fn: Callable, *job_args, **job_kwargs):
        """
        Initialize the runner with a function, its arguments (or dict args),
        and a queue to store the results in the main thread on the main process


        queue
          Queue     queue to store the results in the main process
        job_fn
          Callable  function that will be run in the subprocesses
        job_args
          tuple     Arguments for the function (sent as tuple)
        job_kwargs
          dict      Arguments for the function (sent as dict)

        """

        super().__init__()
        self.MAX_WORKERS = 10
        self.id = uuid4().hex
        self.signals = RunnerSignals()
        self.job_fn = job_fn
        self.job_args = job_args
        self.job_kwargs = job_kwargs

    @pyqtSlot()
    def run(self):
        with concurrent.futures.ProcessPoolExecutor(
            max_workers=self.MAX_WORKERS
        ) as executor:
            job_futures = {
                executor.submit(self.job_fn, *self.job_args, *self.job_kwargs)
                for _ in range(10)
            }

            for job_future in concurrent.futures.as_completed(job_futures):
                result = job_future.result()
                self.signals.results.emit(f"ran job with delay: {result}")

        self.signals.finished.emit(True)
