package websocket

func TransformCursorPosition(cursorPos int, operationType MessageType, operation any, blockID string, userID string) int {
	transformedPos := cursorPos

	switch operationType {
	case MsgInsertChar:
		if insertOp, ok := operation.(*InsertCharOperation); ok {
			if insertOp.BlockID == blockID {
				if insertOp.UserID == userID || insertOp.Position < cursorPos {
					transformedPos++
				}
			}
		}
	case MsgDeleteChar:
		if deleteOp, ok := operation.(*DeleteCharOperation); ok {
			if deleteOp.BlockID == blockID {
				if deleteOp.UserID == userID || deleteOp.Position < cursorPos {
					transformedPos--
				}
			}
		}
	}

	return transformedPos
}
