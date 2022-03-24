import concurrent.futures
import random
import sys
import time
from multiprocessing import Manager, Queue
from processing.listener import Listener
from processing.runner import Runner

from PyQt6.QtCore import QRunnable, QSize, QThreadPool, pyqtSlot
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


# Describe the main Qt windows for the YAL application
class MainWindow(QMainWindow):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        print("MainWindow __init__")
        self.setWindowTitle("My App")
        self.setMinimumSize(QSize(400, 300))

        workers_layout = QVBoxLayout()
        start_button = QPushButton("START IT UP")
        click_me_button = QPushButton("Click me, I'm not frozen")
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
        self.listener = Listener(self.process_queue)
        self.threadpool.start(runner)
        self.threadpool.start(self.listener)

    def update_progress(self, progress):
        self.worker_progress.setValue(progress)


# Main thread
if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = MainWindow()
    app.exec()
