import concurrent.futures
import random
import sys
import time
from typing import Callable
import uuid
from multiprocessing import Manager, Queue

from PyQt6.QtCore import QObject, QRunnable, QSize, QThreadPool, pyqtSignal, pyqtSlot
from PyQt6.QtWidgets import (
    QApplication,
    QLabel,
    QMainWindow,
    QProgressBar,
    QPushButton,
    QVBoxLayout,
    QWidget,
)


def job_fn(queue: Queue, argument1, argument2):
    print(f"Running job_fn with args: {argument1} | {argument2}")
    delay = random.random() * 2  # Random delay value.
    time.sleep(delay)
    queue.put(f"job delay: {delay}")
    return delay


class WorkerSignals(QObject):
    """
    Defines the signals available from a running worker thread.

    progress
        int progress complete,from 0-100

    """

    running = pyqtSignal(str)
    progress = pyqtSignal(str, int)
    finished = pyqtSignal(bool)


class Runner(QRunnable):
    """
    Worker thread
    Inherits from QRunnable to handle worker thread setup, signals and wrap-up.

    """

    def __init__(self, queue: Queue, job_fn: Callable, *job_args, **job_kwargs):
        super().__init__()
        self.MAX_WORKERS = 10
        self.runner_id = uuid.uuid4().hex
        self.signals = WorkerSignals()
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


class Listener(QRunnable):
    """
    Worker thread
    Inherits from QRunnable to handle worker thread setup, signals and wrap-up.

    """

    def __init__(self, queue: Queue):
        super().__init__()
        self.signals = WorkerSignals()
        self.queue = queue

    @pyqtSlot()
    def run(self):
        while True:
            value = self.queue.get()
            if value == "done":
                break
            print(f"Queue value: {value}")
            self.queue.task_done()
        self.signals.finished.emit(True)


# Describe the main Qt windows for the YAL application
class MainWindow(QMainWindow):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        print("MainWindow __init__")
        self.setWindowTitle("My App")
        self.setMinimumSize(QSize(400, 300))

        workers_layout = QVBoxLayout()
        start_button = QPushButton("START IT UP")
        click_me_button = QPushButton("Click me")
        # Progress bar
        self.worker_progress = QProgressBar()
        start_button.pressed.connect(self.execute)
        click_me_button.pressed.connect(self.click_me)
        # Workers running
        self.worker_count = QLabel("0 workers")
        workers_layout.addWidget(self.worker_count)
        workers_layout.addWidget(self.worker_progress)
        workers_layout.addWidget(start_button)
        workers_layout.addWidget(click_me_button)

        # create a queue to communicate with the worker processes
        self.process_queue = Manager().Queue()
        self.threadpool = QThreadPool()
        self.threadpool.setMaxThreadCount(self.threadpool.maxThreadCount() - 5)
        print(f"Starting a threadpool with {self.threadpool.maxThreadCount()} threads")

        mainWidget = QWidget()
        mainWidget.setLayout(workers_layout)
        self.setCentralWidget(mainWidget)
        self.show()

    def click_me(self):
        print("I'm clicked")

    def execute(self):
        # Execute
        runner = Runner(self.process_queue, job_fn, "test1", 2)
        listener = Listener(self.process_queue)
        self.threadpool.start(runner)
        self.threadpool.start(listener)

    def update_progress(self, progress):
        self.worker_progress.setValue(progress)


# Main thread
if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = MainWindow()
    app.exec()
