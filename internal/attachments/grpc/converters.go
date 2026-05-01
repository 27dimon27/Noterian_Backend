package grpc

import (
	attachmentsGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ========== ATTACHMENT CONVERTERS ==========

// ToProtoAttachment конвертирует models.Attachment в protobuf Attachment
func ToProtoAttachment(attachment *models.Attachment) *attachmentsGrpc.Attachment {
	if attachment == nil {
		return nil
	}

	return &attachmentsGrpc.Attachment{
		Id:           attachment.ID.String(),
		BlockId:      attachment.BlockID.String(),
		MinioKey:     attachment.MinioKey,
		AttachUrl:    attachment.AttachURL,
		UrlExpiresAt: timestamppb.New(attachment.URLExpiresAt),
		CreatedAt:    timestamppb.New(attachment.CreatedAt),
		UpdatedAt:    timestamppb.New(attachment.UpdatedAt),
	}
}

// FromProtoAttachment конвертирует protobuf Attachment в models.Attachment
func FromProtoAttachment(proto *attachmentsGrpc.Attachment) *models.Attachment {
	if proto == nil {
		return nil
	}

	return &models.Attachment{
		ID:           uuid.MustParse(proto.GetId()),
		BlockID:      uuid.MustParse(proto.GetBlockId()),
		MinioKey:     proto.GetMinioKey(),
		AttachURL:    proto.GetAttachUrl(),
		URLExpiresAt: proto.GetUrlExpiresAt().AsTime(),
		CreatedAt:    proto.GetCreatedAt().AsTime(),
		UpdatedAt:    proto.GetUpdatedAt().AsTime(),
	}
}

// ========== FILE METADATA CONVERTERS ==========

// AttachmentFileMetadata структура для метаданных файла вложения
type AttachmentFileMetadata struct {
	NoteID   uuid.UUID
	BlockID  uuid.UUID
	FileName string
	FileSize int64
	MimeType string
}

// ToProtoAttachmentFileMetadata конвертирует AttachmentFileMetadata в protobuf FileMetadata
func ToProtoAttachmentFileMetadata(metadata *AttachmentFileMetadata) *attachmentsGrpc.FileMetadata {
	if metadata == nil {
		return nil
	}

	return &attachmentsGrpc.FileMetadata{
		NoteId:   metadata.NoteID.String(),
		BlockId:  metadata.BlockID.String(),
		FileName: metadata.FileName,
		FileSize: metadata.FileSize,
		MimeType: metadata.MimeType,
	}
}

// FromProtoAttachmentFileMetadata конвертирует protobuf FileMetadata в AttachmentFileMetadata
func FromProtoAttachmentFileMetadata(proto *attachmentsGrpc.FileMetadata) *AttachmentFileMetadata {
	if proto == nil {
		return nil
	}

	noteID, _ := uuid.Parse(proto.GetNoteId())
	blockID, _ := uuid.Parse(proto.GetBlockId())

	return &AttachmentFileMetadata{
		NoteID:   noteID,
		BlockID:  blockID,
		FileName: proto.GetFileName(),
		FileSize: proto.GetFileSize(),
		MimeType: proto.GetMimeType(),
	}
}
