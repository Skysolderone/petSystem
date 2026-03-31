import { create } from "zustand";

interface PetState {
  selectedPetId: string | null;
  setSelectedPetId: (petId: string | null) => void;
}

export const usePetStore = create<PetState>((set) => ({
  selectedPetId: null,
  setSelectedPetId: (selectedPetId) => set({ selectedPetId }),
}));
