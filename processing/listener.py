from multiprocessing import Queue
from threading import Event
from PyQt6.QtCore import QObject, QRunnable, pyqtSignal, pyqtSlot


class ListenerSignals(QObject):
    """
    Defines the signals available from a running listener thread.

    running
        bool true if the listener is processing the queue

    """

    running = pyqtSignal(bool)
    finished = pyqtSignal(bool)


class Listener(QRunnable):
    """Listener thread. Inherits from QRunnable to handle worker thread setup, signals and wrap-up."""

    def __init__(self, queue: Queue):
        """
        Initialize the listener thread with a queue to get the results from the main thread.


        queue
          Queue     queue to store the results in the main process

        """
        super().__init__()
        self.signals = ListenerSignals()
        self.close_event = Event()
        self.queue = queue

    @pyqtSlot()
    def run(self):
        """Loop through the queue until we receive a message that indicates the jobs have all been processed"""
        self.signals.running.emit(True)
        while not self.close_event.is_set():
            self.close_event.wait(150)
            value = self.queue.get()
            if value == "queue_done":
                break
            print(f"Queue value: {value}")
            self.queue.task_done()
        print("Listener run(): loop exit")
        self.signals.running.emit(False)
        self.signals.finished.emit(True)

    def start(self):
        self.run()

    def stop(self):
        self.close_event.set()
        print("Listener stop()")
