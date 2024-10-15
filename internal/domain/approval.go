package domain

type RequestStatus string

const (
	RequestStatusApproved RequestStatus = "approved"
	RequestStatusRejected RequestStatus = "rejected"
	RequestStatusPending  RequestStatus = "pending"
	RequestStatusCanceled RequestStatus = "canceled"
)
