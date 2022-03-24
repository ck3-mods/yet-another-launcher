import concurrent.futures
from multiprocessing import Queue
from PyQt6.QtCore import QObject, QRunnable, pyqtSignal, pyqtSlot
from typing import Callable
import uuid


class RunnerSignals(QObject):
    """
    Defines the signals available from a running worker thread.

    progress
        int progress complete,from 0-100

    """

    running = pyqtSignal(str)
    progress = pyqtSignal(str, int)
    finished = pyqtSignal(bool)


class Runner(QRunnable):
    """Runner thread. Inherits from QRunnable to handle runner thread setup, signals and wrap-up."""

    def __init__(self, queue: Queue, job_fn: Callable, *job_args, **job_kwargs):
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
        self.runner_id = uuid.uuid4().hex
        self.signals = RunnerSignals()
        self.job_fn = job_fn
        self.job_args = (queue,) + job_args
        self.job_kwargs = job_kwargs
        self.queue = queue

    @pyqtSlot()
    def run(self):
        with concurrent.futures.ProcessPoolExecutor(
            max_workers=self.MAX_WORKERS
        ) as executor:
            job_futures = {
                executor.submit(self.job_fn, *self.job_args, *self.job_kwargs)
                for _ in range(10)
            }

            for future in job_futures:
                print(f"future: {future}")

        while job_futures:
            done, job_futures = concurrent.futures.wait(
                job_futures, return_when=concurrent.futures.FIRST_COMPLETED
            )
            for job_future in done:
                print(f"future result: {job_future.result()}")
        self.queue.put("done")
        self.signals.finished.emit(True)
