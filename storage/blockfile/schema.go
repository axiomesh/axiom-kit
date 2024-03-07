package blockfile

const (
	// freezerBodiesTable indicates the name of the freezer block body table.
	BlockFileHeaderTable = "header"

	// freezerHeaderTable indicates the name of the freezer header table.
	BlockFileTXsTable = "transactions"

	BlockFileExtraTable = "extra"

	// freezerReceiptTable indicates the name of the freezer receipts table.
	BlockFileReceiptsTable = "receipts"
)

var BlockFileSchema = map[string]bool{
	BlockFileHeaderTable:   true,
	BlockFileTXsTable:      true,
	BlockFileExtraTable:    true,
	BlockFileReceiptsTable: true,
}
