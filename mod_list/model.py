from PyQt6.QtCore import QAbstractListModel, Qt


class ModListModel(QAbstractListModel):
    def __init__(self):
        super().__init__()
        self.mod_list = []

    def data(self, index, role):
        if role == Qt.ItemDataRole.DisplayRole:
            value = self.mod_list[index.row()]
            return value

    def rowCount(self, index):
        return len(self.mod_list)

    def append(self, mod: object):
        self.mod_list.append(mod)
        self.layoutChanged.emit()
