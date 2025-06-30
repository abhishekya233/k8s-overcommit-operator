// SPDX-FileCopyrightText: 2025 2025 INDUSTRIA DE DISEÑO TEXTIL S.A. (INDITEX S.A.)
// SPDX-FileContributor: enriqueavi@inditex.com
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureResourceDeleted(ctx context.Context, c client.Client, obj client.Object) error {
	err := c.Delete(ctx, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}
