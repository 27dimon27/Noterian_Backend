package grpc

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoNote(note *models.Note) *notesGrpc.Note {
	if note == nil {
		return nil
	}

	proto := &notesGrpc.Note{
		Id:        note.ID.String(),
		UserId:    note.UserID.String(),
		Title:     note.Title,
		IsPublic:  note.IsPublic,
		CreatedAt: timestamppb.New(note.CreatedAt),
		UpdatedAt: timestamppb.New(note.UpdatedAt),
	}

	if note.ParentID != nil {
		parentID := note.ParentID.String()
		proto.ParentId = &parentID
	}

	return proto
}

func FromProtoNote(proto *notesGrpc.Note) *models.Note {
	if proto == nil {
		return nil
	}

	note := &models.Note{
		ID:        uuid.MustParse(proto.GetId()),
		UserID:    uuid.MustParse(proto.GetUserId()),
		Title:     proto.GetTitle(),
		IsPublic:  proto.GetIsPublic(),
		CreatedAt: proto.GetCreatedAt().AsTime(),
		UpdatedAt: proto.GetUpdatedAt().AsTime(),
	}

	if proto.ParentId != nil {
		parentID := uuid.MustParse(*proto.ParentId)
		note.ParentID = &parentID
	}

	return note
}

func FromProtoNotes(protos []*notesGrpc.Note) []models.Note {
	notes := make([]models.Note, 0, len(protos))
	for _, proto := range protos {
		if note := FromProtoNote(proto); note != nil {
			notes = append(notes, *note)
		}
	}
	return notes
}

func ToProtoBlock(block *models.Block) *notesGrpc.Block {
	if block == nil {
		return nil
	}

	return &notesGrpc.Block{
		Id:          block.ID.String(),
		NoteId:      block.NoteID.String(),
		BlockTypeId: int32(block.BlockTypeID),
		Position:    int32(block.Position),
		Content:     block.Content,
		CreatedAt:   timestamppb.New(block.CreatedAt),
		UpdatedAt:   timestamppb.New(block.UpdatedAt),
	}
}

func FromProtoBlock(proto *notesGrpc.Block) *models.Block {
	if proto == nil {
		return nil
	}

	return &models.Block{
		ID:          uuid.MustParse(proto.GetId()),
		NoteID:      uuid.MustParse(proto.GetNoteId()),
		BlockTypeID: int(proto.GetBlockTypeId()),
		Position:    int(proto.GetPosition()),
		Content:     proto.GetContent(),
		CreatedAt:   proto.GetCreatedAt().AsTime(),
		UpdatedAt:   proto.GetUpdatedAt().AsTime(),
	}
}

func FromProtoBlocks(protos []*notesGrpc.Block) []models.Block {
	blocks := make([]models.Block, 0, len(protos))
	for _, proto := range protos {
		if block := FromProtoBlock(proto); block != nil {
			blocks = append(blocks, *block)
		}
	}
	return blocks
}

func ToProtoFormattingRange(rng *models.FormattingRange) *notesGrpc.FormattingRange {
	if rng == nil {
		return nil
	}

	proto := &notesGrpc.FormattingRange{
		StartPos: int32(rng.StartPos),
		EndPos:   int32(rng.EndPos),
	}

	if rng.Bold != nil {
		proto.Bold = rng.Bold
	}
	if rng.Italic != nil {
		proto.Italic = rng.Italic
	}
	if rng.Underline != nil {
		proto.Underline = rng.Underline
	}
	if rng.TextAlign != nil {
		align := int32(*rng.TextAlign)
		proto.TextAlign = &align
	}

	return proto
}

func FromProtoFormattingRange(proto *notesGrpc.FormattingRange) *models.FormattingRange {
	if proto == nil {
		return nil
	}

	rng := &models.FormattingRange{
		StartPos: int(proto.GetStartPos()),
		EndPos:   int(proto.GetEndPos()),
	}

	if proto.Bold != nil {
		rng.Bold = proto.Bold
	}
	if proto.Italic != nil {
		rng.Italic = proto.Italic
	}
	if proto.Underline != nil {
		rng.Underline = proto.Underline
	}
	if proto.TextAlign != nil {
		align := int(*proto.TextAlign)
		rng.TextAlign = &align
	}

	return rng
}

func ToProtoBlockFormatting(formatting *models.BlockFormatting) *notesGrpc.BlockFormatting {
	if formatting == nil {
		return nil
	}

	ranges := make([]*notesGrpc.FormattingRange, 0, len(formatting.Ranges))
	for i := range formatting.Ranges {
		ranges = append(ranges, ToProtoFormattingRange(&formatting.Ranges[i]))
	}

	return &notesGrpc.BlockFormatting{
		BlockId: formatting.BlockID,
		Ranges:  ranges,
	}
}

func FromProtoBlockFormatting(proto *notesGrpc.BlockFormatting) *models.BlockFormatting {
	if proto == nil {
		return nil
	}

	ranges := make([]models.FormattingRange, 0, len(proto.GetRanges()))
	for _, rngProto := range proto.GetRanges() {
		if rng := FromProtoFormattingRange(rngProto); rng != nil {
			ranges = append(ranges, *rng)
		}
	}

	return &models.BlockFormatting{
		BlockID: proto.GetBlockId(),
		Ranges:  ranges,
	}
}

func ToProtoBlockWithFormatting(block *models.Block, formatting *models.BlockFormatting) *notesGrpc.BlockWithFormatting {
	return &notesGrpc.BlockWithFormatting{
		Block:      ToProtoBlock(block),
		Formatting: ToProtoBlockFormatting(formatting),
	}
}

func FromProtoBlockWithFormatting(proto *notesGrpc.BlockWithFormatting) (*models.Block, *models.BlockFormatting) {
	if proto == nil {
		return nil, nil
	}
	return FromProtoBlock(proto.GetBlock()), FromProtoBlockFormatting(proto.GetFormatting())
}
