import random
import sys
import time
from queue import Queue

from PyQt6.QtCore import QSize, QThreadPool
from PyQt6.QtWidgets import (
    QApplication,
    QMainWindow,
    QPushButton,
    QVBoxLayout,
    QWidget,
)
from mod_list.model import ModListModel
from mod_list.view import ModList

from processing.runner import Runner


def job_fn(*args, **kwargs):
    delay = random.random() * 2  # Random delay value.
    time.sleep(delay)
    return delay


# Describe the main Qt windows for the YAL application
class MainWindow(QMainWindow):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        print("MainWindow __init__")
        self.setWindowTitle("Yet Another Launcher")
        self.setMinimumSize(QSize(400, 300))

        # create a queue to communicate with the worker processes
        self.process_queue = Queue()
        self.threadpool = QThreadPool()
        self.threadpool.setMaxThreadCount(self.threadpool.maxThreadCount() - 5)
        print(f"Starting a threadpool with {self.threadpool.maxThreadCount()} threads")

        # UI setup
        main_layout = QVBoxLayout()
        # mod list
        mod_list = ModList()
        self.mod_model = ModListModel()
        mod_list.setModel(self.mod_model)
        main_layout.addWidget(mod_list)
        # get mods button
        start_button = QPushButton("Get mods")
        start_button.pressed.connect(self.execute)
        main_layout.addWidget(start_button)
        # click me button
        click_me_button = QPushButton("Click me, I'm not frozen")
        click_me_button.pressed.connect(self.click_me)
        main_layout.addWidget(click_me_button)

        mainWidget = QWidget()
        mainWidget.setLayout(main_layout)
        self.setCentralWidget(mainWidget)
        self.show()

    def closeEvent(self, event):
        print("closeEvent called")
        event.accept()

    def click_me(self):
        self.mod_model.append("I was clicked")

    def execute(self):
        # Execute
        runner = Runner(job_fn, "test1", 2)
        runner.signals.results.connect(self.print_result)
        runner.signals.results.connect(self.mod_model.append)
        self.threadpool.start(runner)

    def print_result(self, result):
        print(f"result: {result}")


# Main thread
if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = MainWindow()
    app.exec()
