package accounting

import "strings"

type TransferService struct {
	repo TransactionRepository
}

func NewTransferService(repo TransactionRepository) *TransferService {
	return &TransferService{repo: repo}
}

func (s *TransferService) Create(userID string, fromInput TransactionCreateInput, toInput TransactionCreateInput) (Transaction, []string, error) {
	pairID := nextID()
	err := s.repo.WithTransferPairLock(userID, pairID, func() error {
		_, createErr := s.repo.CreateTransferPair(userID, pairID, fromInput, toInput)
		return createErr
	})
	if err != nil {
		return Transaction{}, nil, err
	}

	pair, pairErr := s.repo.ListByTransferPairForUser(userID, pairID)
	if pairErr != nil {
		return Transaction{}, nil, pairErr
	}
	if err := validateTransferPair(pair); err != nil {
		return Transaction{}, nil, err
	}

	ledgers := make([]string, 0, 2)
	primary := pair[0]
	for _, side := range pair {
		ledgers = appendUnique(ledgers, side.LedgerID)
		if side.TransferSide == transferSideFrom {
			primary = side
		}
	}

	return primary, ledgers, nil
}

func (s *TransferService) Edit(userID string, txnID string, amount float64, expectedVersion *int) (Transaction, []string, error) {
	pair, err := s.repo.GetTransferPairByTxnID(userID, txnID)
	if err != nil {
		return Transaction{}, nil, err
	}
	if err := validateTransferPair(pair); err != nil {
		return Transaction{}, nil, err
	}

	pairID := strings.TrimSpace(ptrString(pair[0].TransferPairID))
	if pairID == "" {
		return Transaction{}, nil, errTransferBilateralMismatch
	}

	updated := Transaction{}
	lockErr := s.repo.WithTransferPairLock(userID, pairID, func() error {
		next, updateErr := s.repo.UpdateTransferPairAmount(userID, pairID, amount, expectedVersion)
		if updateErr != nil {
			return updateErr
		}
		updated = next
		return nil
	})
	if lockErr != nil {
		return Transaction{}, nil, lockErr
	}

	ledgers := make([]string, 0, 2)
	for _, side := range pair {
		ledgers = appendUnique(ledgers, side.LedgerID)
	}

	return updated, ledgers, nil
}

func (s *TransferService) Delete(userID string, txnID string, expectedVersion *int) ([]string, error) {
	pair, err := s.repo.GetTransferPairByTxnID(userID, txnID)
	if err != nil {
		return nil, err
	}
	if err := validateTransferPair(pair); err != nil {
		return nil, err
	}

	pairID := strings.TrimSpace(ptrString(pair[0].TransferPairID))
	if pairID == "" {
		return nil, errTransferBilateralMismatch
	}

	var ledgers []string
	lockErr := s.repo.WithTransferPairLock(userID, pairID, func() error {
		deletedLedgers, deleteErr := s.repo.DeleteTransferPairByTxnID(userID, txnID, expectedVersion)
		if deleteErr != nil {
			return deleteErr
		}
		ledgers = deletedLedgers
		return nil
	})
	if lockErr != nil {
		return nil, lockErr
	}

	return ledgers, nil
}
